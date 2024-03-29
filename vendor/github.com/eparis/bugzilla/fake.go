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
	"errors"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/util/sets"
)

// Fake is a fake Bugzilla client with injectable fields
type Fake struct {
	EndpointString string
	Bugs           map[int]Bug
	BugErrors      sets.Int
	ExternalBugs   map[int][]ExternalBug
}

func (c *Fake) WithCGIClient(user, password string) Client {
	panic("implement me")
}

func (c *Fake) BugList(queryName, sharerID string) ([]Bug, error) {
	panic("implement me")
}

// Endpoint returns the endpoint for this fake
func (c *Fake) Endpoint() string {
	return c.EndpointString
}

// GetBug retrieves the bug, if registered, or an error, if set,
// or responds with an error that matches IsNotFound
func (c *Fake) GetBug(id int) (*Bug, error) {
	if c.BugErrors.Has(id) {
		return nil, errors.New("injected error getting bug")
	}
	if bug, exists := c.Bugs[id]; exists {
		return &bug, nil
	}
	return nil, &requestError{statusCode: http.StatusNotFound, message: "bug not registered in the fake"}
}

// GetBugComments retrieves the comments of a Bug from the server
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/comment.html#get-comments
func (c *Fake) GetBugComments(id int) ([]Comment, error) {
	panic("implement me")
}

// GetBugHistory retrieves the history of a Bug from the server
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/bug.html#bug-history
func (c *Fake) GetBugHistory(id int) ([]History, error) {
	panic("implement me")
}

// Search doesn't really work, it always returns all bugs
func (c *Fake) Search(query Query) ([]*Bug, error) {
	bugs := []*Bug{}
	for i, _ := range c.Bugs {
		bug := c.Bugs[i]
		bugs = append(bugs, &bug)
	}
	return bugs, nil
}

// GetExternalBugPRsOnBug retrieves the external bugs for the Bugzilla bug,
// if registered, or an error, if set, or responds with an
// error that matches IsNotFound. It filters them by Github PRs.
func (c *Fake) GetExternalBugPRsOnBug(id int) ([]ExternalBug, error) {
	if c.BugErrors.Has(id) {
		return nil, errors.New("injected error adding external bug to bug")
	}
	if _, exists := c.Bugs[id]; exists {
		return c.ExternalBugs[id], nil
	}
	return nil, &requestError{statusCode: http.StatusNotFound, message: "bug not registered in the fake"}
}

// GetExternalBugs retrieves the external bugs for the Bugzilla bug,
// if registered, or an error, if set, or responds with an
// error that matches IsNotFound.
func (c *Fake) GetExternalBugs(id int) ([]ExternalBug, error) {
	if c.BugErrors.Has(id) {
		return nil, errors.New("injected error adding external bug to bug")
	}
	if _, exists := c.Bugs[id]; exists {
		return c.ExternalBugs[id], nil
	}
	return nil, &requestError{statusCode: http.StatusNotFound, message: "bug not registered in the fake"}
}

// UpdateBug updates the bug, if registered, or an error, if set,
// or responds with an error that matches IsNotFound
func (c *Fake) UpdateBug(id int, update BugUpdate) error {
	if c.BugErrors.Has(id) {
		return errors.New("injected error updating bug")
	}
	if bug, exists := c.Bugs[id]; exists {
		bug.Status = update.Status
		bug.Resolution = update.Resolution
		c.Bugs[id] = bug
		return nil
	}
	return &requestError{statusCode: http.StatusNotFound, message: "bug not registered in the fake"}
}

// AddPullRequestAsExternalBug adds an external bug to the Bugzilla bug,
// if registered, or an error, if set, or responds with an error that
// matches IsNotFound
func (c *Fake) AddPullRequestAsExternalBug(id int, org, repo string, num int) (bool, error) {
	if c.BugErrors.Has(id) {
		return false, errors.New("injected error adding external bug to bug")
	}
	if _, exists := c.Bugs[id]; exists {
		pullIdentifier := IdentifierForPull(org, repo, num)
		for _, bug := range c.ExternalBugs[id] {
			if bug.BugzillaBugID == id && bug.ExternalBugID == pullIdentifier {
				return false, nil
			}
		}
		c.ExternalBugs[id] = append(c.ExternalBugs[id], ExternalBug{
			BugzillaBugID: id,
			ExternalBugID: pullIdentifier,
		})
		return true, nil
	}
	return false, &requestError{statusCode: http.StatusNotFound, message: "bug not registered in the fake"}
}

// SetAuthMethod doesn't do anything and you can only set a blank string
func (c *Fake) SetAuthMethod(authMethod string) error {
	if authMethod != "" {
		return fmt.Errorf("Only support blank authmethod")
	}
	return nil
}

// the Fake is a Client
var _ Client = &Fake{}
