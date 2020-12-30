package main

import (
	"encoding/json"
	"io/ioutil"
	log1 "log"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	auth "k8s.io/api/authentication/v1"
)

func main() {
	http.HandleFunc("/authenticate", authenticate)
	http.HandleFunc("/authorize", authorize)
	log1.Fatal(http.ListenAndServe(":3000", nil))
}

func authenticate(res http.ResponseWriter, req *http.Request) {
	logger := log.WithFields(map[string]interface{}{
		"func": "authenticate",
	})

	logger.Info("authenticating user..")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeErrorResp(logger, res, "unable to read request body")
		return
	}
	if len(body) == 0 {
		writeErrorResp(logger, res, "empty request body")
		return
	}

	var tokenReview auth.TokenReview
	if err := json.Unmarshal(body, &tokenReview); err != nil {
		writeErrorResp(logger, res, "invalid token tokenReview request")
		return
	}

	if tokenReview.Spec.Token == "" {
		writeErrorResp(logger, res, "invalid token tokenReview request")
		return
	}
	//
	//  TODO validate with auth provider
	//
	tokenReview.Status = auth.TokenReviewStatus{
		Authenticated: true,
		User: auth.UserInfo{
			Username: "chinna",
			UID:      "1234",
			Groups:   []string{"test", "test1"},
			Extra:    nil,
		},
		Error: "",
	}
	writeDataResp(logger, res, tokenReview)
	return
}

func authorize(res http.ResponseWriter, req *http.Request) {
	logger := log.WithFields(map[string]interface{}{
		"func": "authenticate",
	})

	logger.Info("authenticating user..")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeErrorResp(logger, res, "unable to read request body")
		return
	}
	if len(body) == 0 {
		writeErrorResp(logger, res, "empty request body")
		return
	}

	var tokenReview auth.TokenReview
	if err := json.Unmarshal(body, &tokenReview); err != nil {
		writeErrorResp(logger, res, "invalid token tokenReview request")
		return
	}

	if tokenReview.Spec.Token == "" {
		writeErrorResp(logger, res, "invalid token tokenReview request")
		return
	}
	//
	//  TODO validate with auth provider
	//
	tokenReview.Status = auth.TokenReviewStatus{
		Authenticated: false,
		Error:         "invalid token",
	}
	writeDataResp(logger, res, tokenReview)
	return
}

// writeErrorResp sends error response to client
func writeErrorResp(logger *log.Entry, res http.ResponseWriter, errMsg string) {

	tokenReview := auth.TokenReview{
		Status: auth.TokenReviewStatus{
			Authenticated: false,
			Error:         errMsg,
		},
	}
	logger.Errorf("error: %s", errMsg)
	b, err := json.Marshal(tokenReview)
	if err != nil {
		logger.Errorf("failed to encode response as json: %s", err)
	}
	res.Header().Set("Content-Type", "application/json")
	if _, err := res.Write(b); err != nil {
		logger.Warnf("error sending json to client %s", err.Error())
	}
}

// writeDataResp
func writeDataResp(logger *log.Entry, res http.ResponseWriter, body interface{}) {
	res.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(body)
	if err != nil {
		err = errors.Wrap(err, "error encoding json")
		writeErrorResp(logger, res, err.Error())
		return
	}
	if _, err := res.Write(b); err != nil {
		logger.Warnf("error sending json to client %s", err.Error())
	}
}
