package blockers

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift/bugzilla-tools/pkg/blockerslack/bugutil"
	"github.com/openshift/bugzilla-tools/pkg/blockerslack/config"
	"github.com/openshift/bugzilla-tools/pkg/bugs"
	"github.com/openshift/bugzilla-tools/pkg/teams"

	//"github.com/openshift/bugzilla-tools/pkg/cache"
	"github.com/openshift/bugzilla-tools/pkg/slack"
)

type BlockersReporter struct {
	config config.OperatorConfig

	bugData     *bugs.BugData
	orgData     *teams.OrgData
	slackClient slack.ChannelClient
}

const (
	countHrefFmt    = "%d bugs"
	assigneeHrefFmt = "%d bugs assigned to %s"

	blockerMsgFmt = "It seems there are %s and these bugs are *release blockers*:\nPlease keep eyes on these today!\n"
	triageMsgFmt  = "I found %s which are untriaged\nPlease make sure all bugs have the _Severity_ and _Priority_ field set and do not have the _blocker?_ flag so I can stop bothering you :-)\n"
)

var (
	seriousKeywords = []string{
		"ServiceDeliveryBlocker",
		"TestBlocker",
		"UpgradeBlocker",
	}
)

func NewBlockersReporter(schedule []string, operatorConfig config.OperatorConfig, bugData *bugs.BugData, orgData *teams.OrgData, slackClient slack.ChannelClient, recorder events.Recorder) factory.Controller {
	c := &BlockersReporter{
		config:      operatorConfig,
		bugData:     bugData,
		orgData:     orgData,
		slackClient: slackClient,
	}
	return factory.New().WithSync(c.sync).ResyncSchedule(schedule...).ToController("BlockersReporter", recorder)
}

type triageResult struct {
	who                     string
	bugs                    []int
	seriousKeywordsIDs      map[string][]int
	blockers                []string
	blockerIDs              []int
	needTriage              []string
	needTriageIDs           []int
	needReviewedInSprintIDs []int
	postIDs                 []int
	nonLowIDs               []int
	totalCount              int
	staleCount              int
	priorityCount           map[string]int
	severityCount           map[string]int
}

func getLinkMsg(hrefFmt, msgFmt, who string, bugs []int, args ...string) string {
	hrefText := fmt.Sprintf(hrefFmt, len(bugs), who)
	linkText := makeBugzillaLink(hrefText, bugs)
	fmtArgs := []interface{}{linkText}
	for _, arg := range args {
		fmtArgs = append(fmtArgs, arg)
	}
	message := fmt.Sprintf(msgFmt, fmtArgs...)
	return message
}

func (tr triageResult) getPersonalMessages() []string {
	messages := []string{}
	blockerLen := len(tr.blockers)
	if blockerLen > 0 {
		message := getLinkMsg(assigneeHrefFmt, blockerMsgFmt, tr.who, tr.blockerIDs)
		messages = append(messages, message)
	}

	needTriageLen := len(tr.needTriage)
	if needTriageLen > 0 {
		message := getLinkMsg(assigneeHrefFmt, triageMsgFmt, tr.who, tr.needTriageIDs)
		messages = append(messages, message)
	}

	return messages
}

// Slack has a semi-undocumented limit of about 4000 characters per message.
// https://api.slack.com/changelog/2018-04-truncating-really-long-messages
// Over 4k they will just split it, which is ok except it breaks links, so we break
// it ourselves. If we end up with a single line longer than 4k well, that sucks, get
// fewer bugs and let slack split it and it'll look gross. Not that big of a deal.
func joinMessages(in []string) []string {
	out := []string{}

	cur := []string{}
	curLen := 0

	for _, line := range in {
		lineLen := len(line)
		if curLen+lineLen < 4000 {
			cur = append(cur, line)
			curLen += lineLen
			continue
		}
		text := strings.Join(cur, "\n")
		out = append(out, text)
		cur = []string{line}
		curLen = lineLen
	}
	text := strings.Join(cur, "\n")
	out = append(out, text)
	return out
}

func (tr triageResult) getTeamMessages() []string {
	sortedPrioNames := []string{
		"urgent",
		"high",
		"medium",
		"low",
		"unspecified",
	}
	severityMessages := []string{}
	for _, p := range sortedPrioNames {
		count := tr.severityCount[p]
		if count > 0 {
			severityMessages = append(severityMessages, fmt.Sprintf("%d _%s_", count, p))
		}
	}
	priorityMessages := []string{}
	for _, p := range sortedPrioNames {
		count := tr.priorityCount[p]
		if count > 0 {
			priorityMessages = append(priorityMessages, fmt.Sprintf("%d _%s_", count, p))
		}
	}
	totalCount := tr.totalCount
	href := fmt.Sprintf("%d Bugs", totalCount)
	link := makeBugzillaLink(href, tr.bugs)
	allBugsMsg := fmt.Sprintf("%s Total", link)

	blockerCount := len(tr.blockers)
	href = fmt.Sprintf("%d Release Blockers", blockerCount)
	blockersMsg := makeBugzillaLink(href, tr.blockerIDs)

	needReviewedInSprint := len(tr.needReviewedInSprintIDs)
	href = fmt.Sprintf("%d Bugs Not Reviewed In This Sprint", needReviewedInSprint)
	upcomingMsg := makeBugzillaLink(href, tr.needReviewedInSprintIDs)

	triageCount := len(tr.needTriage)
	href = fmt.Sprintf("%d Untriaged Bugs", triageCount)
	triageMsg := makeBugzillaLink(href, tr.needTriageIDs)

	postCount := len(tr.postIDs)
	href = fmt.Sprintf("%d Bugs in \"POST\"", postCount)
	postMsg := makeBugzillaLink(href, tr.postIDs)

	nonLowCount := len(tr.nonLowIDs)
	href = fmt.Sprintf("%d Bugs formerly known as blockers", nonLowCount)
	nonLowMsg := makeBugzillaLink(href, tr.nonLowIDs)

	lines := []string{
		fmt.Sprintf("\n:bug: *Today's %s OCP Bug Report:* :bug:\n", tr.who),
		fmt.Sprintf("> %s", allBugsMsg),
		fmt.Sprintf("> Bugs Severity Breakdown: %s", strings.Join(severityMessages, ", ")),
		fmt.Sprintf("> Bugs Priority Breakdown: %s", strings.Join(priorityMessages, ", ")),
		fmt.Sprintf("> %s", blockersMsg),
	}
	if nonLowCount > 0 {
		lines = append(lines, fmt.Sprintf("> %s", nonLowMsg))
	}
	if needReviewedInSprint > 0 {
		lines = append(lines, fmt.Sprintf("> %s", upcomingMsg))
	}
	if triageCount > 0 {
		lines = append(lines, fmt.Sprintf("> %s", triageMsg))
	}
	if postCount > 0 {
		lines = append(lines, fmt.Sprintf("> %s", postMsg))
	}

	if tr.seriousKeywordsIDs != nil {
		for _, keyword := range seriousKeywords {
			if bugIDs, ok := tr.seriousKeywordsIDs[keyword]; ok {
				href := fmt.Sprintf("%d Bugs with %s", len(bugIDs), keyword)
				lines = append(lines, fmt.Sprintf("> %s", makeBugzillaLink(href, bugIDs)))
			}
		}
	}
	lines = joinMessages(lines)

	return lines
}

func triageBug(who string, bugs ...*bugs.Bug) triageResult {
	r := triageResult{
		who:           who,
		totalCount:    len(bugs),
		priorityCount: map[string]int{},
		severityCount: map[string]int{},
	}
	r.bugs = make([]int, 0, len(bugs))
	for _, bug := range bugs {
		r.bugs = append(r.bugs, bug.ID)

		keywords := sets.NewString(bug.Keywords...)
		for _, keyword := range seriousKeywords {
			if keywords.Has(keyword) {
				if r.seriousKeywordsIDs == nil {
					r.seriousKeywordsIDs = make(map[string][]int)
				}
				r.seriousKeywordsIDs[keyword] = append(r.seriousKeywordsIDs[keyword], bug.ID)
			}
		}

		if strings.Contains(bug.Whiteboard, "LifecycleStale") {
			r.staleCount++
			continue
		}

		r.severityCount[bug.Severity]++
		r.priorityCount[bug.Priority]++

		if !bug.ReviewedInSprint() {
			r.needReviewedInSprintIDs = append(r.needReviewedInSprintIDs, bug.ID)
		}

		if bug.Untriaged() {
			r.needTriage = append(r.needTriage, bugutil.FormatBugMessage(bug))
			r.needTriageIDs = append(r.needTriageIDs, bug.ID)
		}

		if bug.Blocker() {
			r.blockers = append(r.blockers, bugutil.FormatBugMessage(bug))
			r.blockerIDs = append(r.blockerIDs, bug.ID)
		}

		if bug.Status == "POST" {
			r.postIDs = append(r.postIDs, bug.ID)
		}

		if !bug.LowPriorityAndSeverity() {
			r.nonLowIDs = append(r.nonLowIDs, bug.ID)
		}
	}

	return r
}

type notificationMap map[string]triageResult

func (c *BlockersReporter) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	if c.config.Debug {
		fmt.Println("Started sync()")
	}
	c.orgData.Reconcile()
	if err := c.bugData.Reconcile(); err != nil {
		return err
	}

	c.bugData = c.bugData.FilterByStatus(bugs.OnEngineeringStatus())

	peopleNotificationMap, teamNotificationMap := Report(ctx, c.orgData, c.bugData, syncCtx.Recorder(), &c.config)

	for person, results := range peopleNotificationMap {
		messages := results.getPersonalMessages()
		if len(messages) == 0 {
			continue
		}
		message := strings.Join(messages, "\n")
		if !c.config.Debug {
			if err := c.slackClient.MessageEmail(person, message); err != nil {
				syncCtx.Recorder().Warningf("DeliveryFailed", "To: %s: %v", person, err)
			}
		}
	}

	notSentToTeam := sets.NewString(c.orgData.GetTeamNames()...)
	sentToTeam := []string{}
	for team, results := range teamNotificationMap {
		if results.totalCount == 0 {
			continue
		}
		teamInfo, ok := c.orgData.Teams[team]
		if !ok {
			syncCtx.Recorder().Warningf("Unable to find team data", "team %q not found", team)
			continue
		}
		slackChan := teamInfo.SlackChan
		if slackChan == "" {
			// If we don't know where to send this team's info, do nothing.
			syncCtx.Recorder().Warningf("Unable to find channel", "team %q not found", team)
			continue
		}
		notSentToTeam.Delete(team)
		sentToTeam = append(sentToTeam, team)
		messages := results.getTeamMessages()
		for _, message := range messages {
			if err := c.slackClient.MessageChannel(slackChan, message); err != nil {
				syncCtx.Recorder().Warningf("DeliveryFailed", "Failed to deliver stats to channel %q: %v", slackChan, err)
			}
		}
	}

	teamMessage := fmt.Sprintf("Sent to team: %s", strings.Join(sentToTeam, ", "))
	notTeamMessage := fmt.Sprintf("Not sent to team: %s", strings.Join(notSentToTeam.List(), ", "))
	messages := []string{teamMessage, notTeamMessage}
	message := strings.Join(messages, "\n\n")
	if err := c.slackClient.MessageDebug(message); err != nil {
		syncCtx.Recorder().Warningf("DeliveryFailed", "Failed to deliver stats to debug channel: %v", err)
	}
	if c.config.Debug {
		os.Exit(0)
	}
	return nil
}

func Report(ctx context.Context, orgData *teams.OrgData, bugData *bugs.BugData, recorder events.Recorder, config *config.OperatorConfig) (peopleNotificationMap notificationMap, teamNotificationMap notificationMap) {
	teamsWithChannel := []string{}
	for team, teamInfo := range orgData.Teams {
		if teamInfo.SlackChan != "" {
			teamsWithChannel = append(teamsWithChannel, team)
		}
	}
	bugData = bugData.FilterByTeams(teamsWithChannel)

	peopleNotificationMap = notificationMap{}
	peopleBugsMap := bugData.GetPeopleMap()
	for person, bugList := range peopleBugsMap {
		result := triageBug(person, bugList...)
		peopleNotificationMap[person] = result
	}

	teamNotificationMap = notificationMap{}
	teamBugsMap := bugData.GetTeamMap()
	for team, bugList := range teamBugsMap {
		result := triageBug(team, bugList...)
		teamNotificationMap[team] = result
	}

	return peopleNotificationMap, teamNotificationMap
}

func makeBugzillaLink(hrefText string, ids []int) string {
	u, _ := url.Parse("https://bugzilla.redhat.com/buglist.cgi")
	e := u.Query()
	e.Add("f1", "bug_id")
	e.Add("o1", "anyexact")
	stringIds := make([]string, len(ids))
	for i := range stringIds {
		stringIds[i] = fmt.Sprintf("%d", ids[i])
	}
	e.Add("v1", strings.Join(stringIds, ","))
	u.RawQuery = e.Encode()
	return fmt.Sprintf("<%s|%s>", u.String(), hrefText)
}
