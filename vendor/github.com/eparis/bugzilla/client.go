/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bugzilla

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	methodField         = "method"
	AuthBearer          = "bearer"
	AuthQuery           = "query"
	AuthXBugzillaAPIKey = "x-bugzilla-api-key"
)

type Client interface {
	Endpoint() string
	GetBug(id int) (*Bug, error)
	GetBugComments(id int) ([]Comment, error)
	GetBugHistory(id int) ([]History, error)
	Search(query Query) ([]*Bug, error)
	GetExternalBugs(id int) ([]ExternalBug, error)
	GetExternalBugPRsOnBug(id int) ([]ExternalBug, error)
	UpdateBug(id int, update BugUpdate) error
	AddPullRequestAsExternalBug(id int, org, repo string, num int) (bool, error)
	SetAuthMethod(authMethod string) error

	WithCGIClient(user, password string) Client
	// only supported with CGI client
	BugList(queryName, sharerID string) ([]Bug, error)
}

func NewClient(getAPIKey func() []byte, endpoint string) Client {
	return &client{
		logger:    logrus.WithField("client", "bugzilla"),
		client:    &http.Client{},
		endpoint:  endpoint,
		getAPIKey: getAPIKey,
	}
}

type client struct {
	logger     *logrus.Entry
	client     *http.Client
	cgiClient  *bugzillaCGIClient
	endpoint   string
	getAPIKey  func() []byte
	authMethod string
}

// the client is a Client impl
var _ Client = &client{}

func (c *client) SetAuthMethod(authMethod string) error {
	if authMethod != "" && authMethod != AuthBearer && authMethod != AuthQuery && authMethod != AuthXBugzillaAPIKey {
		return fmt.Errorf("invalid auth-method %s. Valid values are bearer,query or x-bugzilla-api-key", authMethod)
	}
	c.authMethod = authMethod
	return nil
}

func (c *client) Endpoint() string {
	return c.endpoint
}

func (c *client) WithCGIClient(username, password string) Client {
	var err error
	c.cgiClient, err = newCGIClient(c.endpoint, username, password)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *client) getBugs(url string, values *url.Values, logger *logrus.Entry) ([]*Bug, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if values != nil {
		req.URL.RawQuery = values.Encode()
	}
	raw, err := c.request(req, logger)
	if err != nil {
		return nil, err
	}
	var parsedResponse struct {
		Bugs []*Bug `json:"bugs,omitempty"`
	}
	if err := json.Unmarshal(raw, &parsedResponse); err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %v", err)
	}
	return parsedResponse.Bugs, nil
}

// GetBug retrieves a Bug from the server
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/bug.html#get-bug
func (c *client) GetBug(id int) (*Bug, error) {
	logger := c.logger.WithFields(logrus.Fields{methodField: "GetBug", "id": id})
	url := fmt.Sprintf("%s/rest/bug/%d", c.endpoint, id)
	bugs, err := c.getBugs(url, nil, logger)
	if err != nil {
		return nil, err
	}
	if len(bugs) != 1 {
		return nil, fmt.Errorf("did not get one bug, but %d: %v", len(bugs), bugs)
	}
	return bugs[0], nil
}

// GetBugComments retrieves the comments of a Bug from the server
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/comment.html#get-comments
func (c *client) GetBugComments(id int) ([]Comment, error) {
	logger := c.logger.WithFields(logrus.Fields{methodField: "GetBugComments", "id": id})
	url := fmt.Sprintf("%s/rest/bug/%d/comment", c.endpoint, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	raw, err := c.request(req, logger)
	if err != nil {
		return nil, err
	}
	var parsedResponse struct {
		Bugs map[string]*struct {
			Comments []Comment `json:"comments,omitempty"`
		} `json:"bugs,omitempty"`
	}
	if err := json.Unmarshal(raw, &parsedResponse); err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %v", err)
	}
	if len(parsedResponse.Bugs) != 1 {
		return nil, fmt.Errorf("did not get one bug, but %d: %v", len(parsedResponse.Bugs), parsedResponse.Bugs)
	}
	for _, comments := range parsedResponse.Bugs {
		return comments.Comments, nil
	}
	return nil, nil
}

// GetBugHistory retrieves the history of a Bug from the server
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/bug.html#bug-history
func (c *client) GetBugHistory(id int) ([]History, error) {
	logger := c.logger.WithFields(logrus.Fields{methodField: "GetBugHistory", "id": id})
	url := fmt.Sprintf("%s/rest/bug/%d/history", c.endpoint, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	raw, err := c.request(req, logger)
	if err != nil {
		return nil, err
	}
	var parsedResponse struct {
		Bugs []*struct {
			History []History `json:"history,omitempty"`
		} `json:"bugs,omitempty"`
	}
	if err := json.Unmarshal(raw, &parsedResponse); err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %v", err)
	}
	if len(parsedResponse.Bugs) != 1 {
		return nil, fmt.Errorf("did not get one bug, but %d: %v", len(parsedResponse.Bugs), parsedResponse.Bugs)
	}
	return parsedResponse.Bugs[0].History, nil
}

// GetExternalBugPRsOnBug retrieves external bugs on a Bug from the server
// and returns any that reference a Pull Request in GitHub
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/bug.html#get-bug
func (c *client) GetExternalBugPRsOnBug(id int) ([]ExternalBug, error) {
	ebs, err := c.GetExternalBugs(id)
	if err != nil {
		return nil, err
	}

	return filterPRs(ebs)
}

// GetExternalBugPRsOnBug retrieves external bugs on a Bug from the server
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/bug.html#get-bug
func (c *client) GetExternalBugs(id int) ([]ExternalBug, error) {
	logger := c.logger.WithFields(logrus.Fields{methodField: "GetExternalBugPRsOnBug", "id": id})
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/rest/bug/%d", c.endpoint, id), nil)
	if err != nil {
		return nil, err
	}
	values := req.URL.Query()
	values.Add("include_fields", "external_bugs")
	req.URL.RawQuery = values.Encode()
	raw, err := c.request(req, logger)
	if err != nil {
		return nil, err
	}
	var parsedResponse struct {
		Bugs []struct {
			ExternalBugs []ExternalBug `json:"external_bugs"`
		} `json:"bugs"`
	}
	if err := json.Unmarshal(raw, &parsedResponse); err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %v", err)
	}
	if len(parsedResponse.Bugs) != 1 {
		return nil, fmt.Errorf("did not get one bug, but %d: %v", len(parsedResponse.Bugs), parsedResponse)
	}
	var prs []ExternalBug
	for _, bug := range parsedResponse.Bugs[0].ExternalBugs {
		if bug.BugzillaBugID != id {
			continue
		}
		prs = append(prs, bug)
	}
	return prs, nil
}

func filterPRs(ebs []ExternalBug) ([]ExternalBug, error) {
	var prs []ExternalBug
	for _, bug := range ebs {
		if bug.Type.URL != "https://github.com/" {
			// TODO: skuznets: figure out how to honor the endpoints given to the GitHub client to support enterprise here
			continue
		}
		org, repo, num, err := PullFromIdentifier(bug.ExternalBugID)
		if IsIdentifierNotForPullErr(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("could not parse external identifier %q as pull: %v", bug.ExternalBugID, err)
		}
		bug.Org = org
		bug.Repo = repo
		bug.Num = num
		prs = append(prs, bug)
	}
	return prs, nil
}

// UpdateBug updates the fields of a bug on the server
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/bug.html#update-bug
func (c *client) UpdateBug(id int, update BugUpdate) error {
	body, err := json.Marshal(update)
	logger := c.logger.WithFields(logrus.Fields{methodField: "UpdateBug", "id": id, "update": string(body)})
	if err != nil {
		return fmt.Errorf("failed to marshal update payload: %v", err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/rest/bug/%d", c.endpoint, id), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	_, err = c.request(req, logger)
	return err
}

func (c *client) request(req *http.Request, logger *logrus.Entry) ([]byte, error) {
	logger = logger.WithField("url", obfuscatedURL(req.URL.String())).WithField("verb", req.Method)
	if apiKey := c.getAPIKey(); len(apiKey) > 0 {
		switch c.authMethod {
		case AuthBearer:
			req.Header.Set("Authorization", "Bearer "+string(apiKey))
		case AuthQuery:
			values := req.URL.Query()
			values.Add("api_key", string(apiKey))
			req.URL.RawQuery = values.Encode()
		case AuthXBugzillaAPIKey:
			req.Header.Set("X-BUGZILLA-API-KEY", string(apiKey))
		default:
			// If there is no auth method specified, we use a union of `query` and
			// `x-bugzilla-api-key` to mimic the previous default behavior which attempted
			// to satisfy different BugZilla server versions.
			req.Header.Set("X-BUGZILLA-API-KEY", string(apiKey))
			values := req.URL.Query()
			values.Add("api_key", string(apiKey))
			req.URL.RawQuery = values.Encode()
		}
	}
	start := time.Now()
	resp, err := c.client.Do(req)
	stop := time.Now()
	promLabels := prometheus.Labels(map[string]string{methodField: logger.Data[methodField].(string), "status": ""})
	if resp != nil {
		promLabels["status"] = strconv.Itoa(resp.StatusCode)
	}
	requestDurations.With(promLabels).Observe(float64(stop.Sub(start).Seconds()))
	if resp != nil {
		logger.WithField("response", resp.StatusCode).Debug("Got response from Bugzilla.")
	}
	if err != nil {
		code := -1
		if resp != nil {
			code = resp.StatusCode
		}
		return nil, &requestError{statusCode: code, message: err.Error()}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.WithError(err).Warn("could not close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, &requestError{statusCode: resp.StatusCode, message: fmt.Sprintf("response code %d not %d", resp.StatusCode, http.StatusOK)}
	}
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %v", err)
	}
	return raw, nil
}

type requestError struct {
	statusCode int
	message    string
}

func (e requestError) Error() string {
	return e.message
}

func IsNotFound(err error) bool {
	reqError, ok := err.(*requestError)
	if !ok {
		return false
	}
	return reqError.statusCode == http.StatusNotFound
}

// AddPullRequestAsExternalBug attempts to add a PR to the external tracker list.
// External bugs are assumed to fall under the type identified by their hostname,
// so we will provide https://github.com/ here for the URL identifier. We return
// any error as well as whether a change was actually made.
// This will be done via JSONRPC:
// https://bugzilla.redhat.com/docs/en/html/integrating/api/Bugzilla/Extension/ExternalBugs/WebService.html#add-external-bug
func (c *client) AddPullRequestAsExternalBug(id int, org, repo string, num int) (bool, error) {
	logger := c.logger.WithFields(logrus.Fields{methodField: "AddExternalBug", "id": id, "org": org, "repo": repo, "num": num})
	pullIdentifier := IdentifierForPull(org, repo, num)
	rpcPayload := struct {
		// Version is the version of JSONRPC to use. All Bugzilla servers
		// support 1.0. Some support 1.1 and some support 2.0
		Version string `json:"jsonrpc"`
		Method  string `json:"method"`
		// Parameters must be specified in JSONRPC 1.0 as a structure in the first
		// index of this slice
		Parameters []AddExternalBugParameters `json:"params"`
		ID         string                     `json:"id"`
	}{
		Version: "1.0", // some Bugzilla servers support 2.0 but all support 1.0
		Method:  "ExternalBugs.add_external_bug",
		ID:      "identifier", // this is useful when fielding asynchronous responses, but not here
		Parameters: []AddExternalBugParameters{{
			APIKey: string(c.getAPIKey()),
			BugIDs: []int{id},
			ExternalBugs: []NewExternalBugIdentifier{{
				Type: "https://github.com/",
				ID:   pullIdentifier,
			}},
		}},
	}
	body, err := json.Marshal(rpcPayload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal JSONRPC payload: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/jsonrpc.cgi", c.endpoint), bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.request(req, logger)
	if err != nil {
		return false, err
	}
	var response struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
		ID     string `json:"id"`
		Result *struct {
			Bugs []struct {
				ID      int `json:"id"`
				Changes struct {
					ExternalBugs struct {
						Added   string `json:"added"`
						Removed string `json:"removed"`
					} `json:"ext_bz_bug_map.ext_bz_bug_id"`
				} `json:"changes"`
			} `json:"bugs"`
		} `json:"result,omitempty"`
	}
	if err := json.Unmarshal(resp, &response); err != nil {
		return false, fmt.Errorf("failed to unmarshal JSONRPC response: %v", err)
	}
	if response.Error != nil {
		if response.Error.Code == 100500 && strings.Contains(response.Error.Message, `duplicate key value violates unique constraint "ext_bz_bug_map_bug_id_idx"`) {
			// adding the external bug failed since it is already added, this is not an error
			return false, nil
		}
		return false, fmt.Errorf("JSONRPC error %d: %v", response.Error.Code, response.Error.Message)
	}
	if response.ID != rpcPayload.ID {
		return false, fmt.Errorf("JSONRPC returned mismatched identifier, expected %s but got %s", rpcPayload.ID, response.ID)
	}
	changed := false
	if response.Result != nil {
		for _, bug := range response.Result.Bugs {
			if bug.ID == id {
				changed = changed || strings.Contains(bug.Changes.ExternalBugs.Added, pullIdentifier)
			}
		}
	}
	return changed, nil
}

func IdentifierForPull(org, repo string, num int) string {
	return fmt.Sprintf("%s/%s/pull/%d", org, repo, num)
}

func PullFromIdentifier(identifier string) (org, repo string, num int, err error) {
	parts := strings.Split(identifier, "/")
	if len(parts) != 4 {
		return "", "", 0, fmt.Errorf("invalid pull identifier with %d parts: %q", len(parts), identifier)
	}
	if parts[2] != "pull" {
		return "", "", 0, &identifierNotForPull{identifier: identifier}
	}
	number, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid pull identifier: could not parse %s as number: %v", parts[3], err)
	}

	return parts[0], parts[1], number, nil
}

type identifierNotForPull struct {
	identifier string
}

func (i identifierNotForPull) Error() string {
	return fmt.Sprintf("identifier %q is not for a pull request", i.identifier)
}

func IsIdentifierNotForPullErr(err error) bool {
	_, ok := err.(*identifierNotForPull)
	return ok
}

var re = regexp.MustCompile(`api_key=[^&]*&`)

func obfuscatedURL(url string) string {
	return re.ReplaceAllString(url, `api_key=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`)
}
