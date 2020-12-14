package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	authentication "k8s.io/api/authentication/v1"
)

func unauthorized(w http.ResponseWriter, format string, args ...interface{}) {
	log.Printf(format, args...)
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "authentication.k8s.io/v1beta1",
		"kind":       "TokenReview",
		"status": authentication.TokenReviewStatus{
			Authenticated: false,
		},
	})
}
func getGroups(ctx context.Context, client *github.Client, user *github.User, org string, orgMember bool) ([]string, error) {
	opt := &github.ListOptions{PerPage: 1000}
	// If orgMember is set, user must be member of the org
	var groups []string
	if org != "" {
		if orgMember {
			isMember := false
			// Also set large pagination on ListMembers() call
			members, _, err := client.Organizations.ListMembers(ctx, org, &github.ListMembersOptions{ListOptions: *opt})
			if err == nil {
				for _, member := range members {
					if strings.EqualFold(*member.Login, *user.Login) {
						isMember = true
					}
				}
			} else {
				return nil, fmt.Errorf("[Error] getting organization=%s members: %s", org, err.Error())
			}
			if !isMember {
				return nil, fmt.Errorf("[Error] user=%q not in organization=%s", *user.Login, org)
			}
		}
		// Gather teams
		teamsResults := []*github.Team{}
		teamsResults, _, err := client.Teams.ListUserTeams(ctx, opt)
		// Soft fail on listing user's teams (e.g. user with no team at all)
		if err != nil {
			log.Printf("[Warning] failed to list teams for user=%s: %s", *user.Login, err.Error())
		}
		for _, team := range teamsResults {
			if !(strings.EqualFold(org, *team.Organization.Login)) {
				continue
			}
			groups = append(groups, *team.Name)
		}
	}
	return groups, nil
}
func main() {
	org := os.Getenv("GITHUB_ORG")
	orgMember := os.Getenv("GITHUB_ORG_MEMBER")
	http.HandleFunc("/authenticate", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var tr authentication.TokenReview
		err := decoder.Decode(&tr)
		if err != nil {
			unauthorized(w, "[Error] decoding request: %s", err.Error())
			return
		}

		// Check User
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: tr.Spec.Token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client := github.NewClient(tc)
		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			unauthorized(w, "[Error] invalid token: %s", err.Error())
			return
		}

		groups, err := getGroups(ctx, client, user, org, orgMember != "")
		if err != nil {
			unauthorized(w, err.Error())
			return
		}
		// Set the TokenReviewStatus
		log.Printf("[Success] login as user=%s, groups=%v, org=%s", *user.Login, groups, org)
		w.WriteHeader(http.StatusOK)
		trs := authentication.TokenReviewStatus{
			Authenticated: true,
			User: authentication.UserInfo{
				Username: *user.Login,
				UID:      *user.Login,
				Groups:   groups,
			},
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"apiVersion": "authentication.k8s.io/v1",
			"kind":       "TokenReview",
			"status":     trs,
		})
	})
	log.Fatal(http.ListenAndServe(":3000", nil))
}
