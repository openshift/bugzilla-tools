package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/bugzilla-tools/pkg/bugs"
	"github.com/openshift/bugzilla-tools/pkg/teams"
	"github.com/eparis/goSmartSheet"
	//"github.com/kr/pretty"
	"github.com/spf13/cobra"
)

const (
	teamNameColumnName   = "Team Name"
	allBugColumnName     = "Bug Count (All)"
	currentBugColumnName = "Bug Count (Current Release)"

	ssAPIKeyFlagName   = "smartsheet-key"
	ssAPIKeyFlagDefVal = "smartsheetKey"
	ssAPIKeyFlagUsage  = "Path to file containing SmartSheet API key"

	url     = "https://api.smartsheet.com/2.0"
	sheetID = "6386356843767684" // production sheet
	//sheetID = "298546583889796" // eparis sheet
)

var (
	targets = []string{"---", "4.6.0"}
)

func newIntCell(column int64, val int) goSmartSheet.Cell {
	return goSmartSheet.Cell{
		ColumnID: column,
		Value: &goSmartSheet.CellValue{
			IntVal: &val,
		},
	}
}

func getAuthToken(cmd *cobra.Command) (string, error) {
	keyFile, err := cmd.Flags().GetString(ssAPIKeyFlagName)
	dat, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return "", err
	}
	apikey := strings.TrimRight(string(dat), "\r\n")
	return apikey, nil
}

func doMain(cmd *cobra.Command, _ []string) error {
	orgData, err := teams.GetOrgData(cmd)
	if err != nil {
		return err
	}

	bugData, err := bugs.GetBugData(cmd, orgData)
	if err != nil {
		return err
	}
	bugMap := bugData.GetTeamMap()

	ssToken, err := getAuthToken(cmd)
	if err != nil {
		return err
	}
	client, err := goSmartSheet.GetClient(ssToken, url)
	if err != nil {
		return err
	}
	sheet, err := client.GetSheet(sheetID, "")
	if err != nil {
		return err
	}
	var teamNameColumn int64
	var allBugColumn int64
	var currentBugColumn int64
	for _, column := range sheet.Columns {
		switch column.Title {
		case teamNameColumnName:
			teamNameColumn = column.ID
		case allBugColumnName:
			allBugColumn = column.ID
		case currentBugColumnName:
			currentBugColumn = column.ID
		}
	}
	newRows := []goSmartSheet.Row{}
	teamOnSheet := map[string]bool{}
	for _, row := range sheet.Rows {
		for _, cell := range row.Cells {
			if cell.ColumnID != teamNameColumn {
				continue
			}
			if cell.Value == nil || cell.Value.StringVal == nil {
				continue
			}
			teamName := cell.Value.StringVal
			teamOnSheet[*teamName] = true
			_, ok := bugMap[*teamName]
			if !ok {
				fmt.Printf("Unable to find bugs for: %s\n", *teamName)
				continue
			}
			newCells := []goSmartSheet.Cell{
				newIntCell(currentBugColumn, bugMap.CountBlocker(*teamName, targets)),
				newIntCell(allBugColumn, bugMap.CountAll(*teamName)),
			}
			newRow := goSmartSheet.Row{
				ID:    row.ID,
				Cells: newCells,
			}
			newRows = append(newRows, newRow)
		}
	}
	for teamName := range bugMap {
		if !teamOnSheet[teamName] {
			fmt.Printf("Found bugs for %s but not on SoS sheet\n", teamName)
		}
	}
	closer, err := client.UpdateRowsOnSheet(sheetID, newRows)
	if err != nil {
		return err
	}
	_ = closer
	return nil
}

func main() {
	cmd := &cobra.Command{
		Use:  filepath.Base(os.Args[0]),
		RunE: doMain,
	}
	teams.AddFlags(cmd)
	bugs.AddFlags(cmd)
	cmd.Flags().String(ssAPIKeyFlagName, ssAPIKeyFlagDefVal, ssAPIKeyFlagUsage)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
