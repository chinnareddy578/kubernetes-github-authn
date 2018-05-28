package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	authentication "k8s.io/client-go/pkg/apis/authentication/v1beta1"
)

func unauthorized(w http.ResponseWriter, format string, args ...interface{}) {
	log.Printf(format, args...)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"apiVersion": "authentication.k8s.io/v1beta1",
		"kind":       "TokenReview",
		"status": authentication.TokenReviewStatus{
			Authenticated: false,
		},
	})
}
func main() {
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

		opt := &github.ListOptions{PerPage: 1000}
		// If $GITHUB_ORG_MEMBER is set, user must be member of the $GITHUB_ORG organization
		org := os.Getenv("GITHUB_ORG")
		member_check := os.Getenv("GITHUB_ORG_MEMBER")
		if org != "" && member_check != "" {
			is_member := false
			members, _, err := client.Organizations.ListMembers(ctx, org, nil)
			if err == nil {
				for _, member := range members {
					if member.Login != nil && *member.Login == *user.Login {
						is_member = true
					}
				}
			} else {
				unauthorized(w, "[Error] getting organization=%s members: %s", org, err.Error())
				return
			}
			if !is_member {
				unauthorized(w, "[Error] user=%q not in organization=%s", *user.Login, org)
				return
			}
		}

		// Gather teams
		teams_results := []*github.Team{}
		teams_results, _, err = client.Organizations.ListUserTeams(ctx, opt)
		// Soft fail on listing user's teams (e.g. user with no team at all)
		if err != nil {
			log.Printf("[Warning] failed to list teams for user=%s: %s", *user.Login, err.Error())
		}
		var groups []string
		for _, team := range teams_results {
			if org != "" {
				if !(strings.EqualFold(org, *team.Organization.Login)) {
					continue
				}
			}
			groups = append(groups, *team.Name)
		}

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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"apiVersion": "authentication.k8s.io/v1beta1",
			"kind":       "TokenReview",
			"status":     trs,
		})
	})
	log.Fatal(http.ListenAndServe(":3000", nil))
}
