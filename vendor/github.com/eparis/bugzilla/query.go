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
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// Values returns a url.Values strcture based on the query search parameters.
func (q *Query) Values() *url.Values {
	values := &url.Values{}
	for _, val := range q.Classification {
		values.Add("classification", val)
	}
	for _, val := range q.Product {
		values.Add("product", val)
	}
	for _, val := range q.Status {
		values.Add("bug_status", val)
	}
	for _, val := range q.Priority {
		values.Add("priority", val)
	}
	for _, val := range q.Severity {
		values.Add("bug_severity", val)
	}
	for _, val := range q.Component {
		values.Add("component", val)
	}
	for _, val := range q.Keywords {
		values.Add("keywords", val)
		if q.KeywordsType == "" {
			panic("Invalid query: Keyworrds set but KeywordsType unset")
		}
	}
	if q.KeywordsType != "" {
		values.Add("keywords_type", q.KeywordsType)
	}
	for _, val := range q.BugIDs {
		values.Add("bug_id", val)
		if q.BugIDsType == "" {
			panic("Invalid query: BugIDs set but BugIDsType unset")
		}
	}
	if q.BugIDsType != "" {
		values.Add("bug_id_type", q.BugIDsType)
	}
	for _, val := range q.TargetRelease {
		values.Add("target_release", val)
	}
	for i, adv := range q.Advanced {
		fieldNum := i + 1
		values.Set(fmt.Sprintf("f%d", fieldNum), adv.Field)
		values.Set(fmt.Sprintf("o%d", fieldNum), adv.Op)
		if adv.Value != "" {
			values.Set(fmt.Sprintf("v%d", fieldNum), adv.Value)
		}
		if adv.Negate {
			values.Set(fmt.Sprintf("n%d", fieldNum), "1")
		}
	}
	if len(q.IncludeFields) != 0 {
		fields := strings.Join(q.IncludeFields, ",")
		values.Set("include_fields", fields)
	}
	v, err := url.ParseQuery(q.Raw)
	if err != nil {
		logrus.Warnf("Unable to parse Raw search query: %q: %v", q.Raw, err)
	}
	for k, vals := range v {
		for _, val := range vals {
			values.Add(k, val)
		}
	}
	return values
}

// Search retrieves all Bugs matching the search
// https://bugzilla.readthedocs.io/en/latest/api/core/v1/bug.html#search-bugs
func (c *client) Search(query Query) ([]*Bug, error) {
	limit := 0
	offset := 0
	outbugs := []*Bug{}

	logger := c.logger.WithFields(logrus.Fields{methodField: "Search"})
	url := fmt.Sprintf("%s/rest/bug", c.endpoint)

	values := query.Values()
	for {
		values.Set("limit", fmt.Sprint(limit))
		values.Set("offset", fmt.Sprint(offset))
		bugs, err := c.getBugs(url, values, logger)
		if err != nil {
			return nil, err
		}
		if len(bugs) == 0 {
			break
		}
		outbugs = append(outbugs, bugs...)

		// If we do a query and get back N bugs we assume that N was the maximum number of bugs we can get
		// If the server can send us 1,000 bugs and we get back only 12, we're going to assume that 12 was
		// the server limit. And we are going to do a second query with limit = 12, offset = 12. That second
		// query will return 0 bugs and we will break.
		// That wasted second query wouldn't be needed if we could tell how many total bugs existed or if we
		// knew the server limit. Since we don't have either, best we can do it guess and test.
		if limit == 0 {
			limit = len(bugs)
		}
		if len(bugs) < limit {
			break
		}
		offset += limit
	}
	return outbugs, nil
}
