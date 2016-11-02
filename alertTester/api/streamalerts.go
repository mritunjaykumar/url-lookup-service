package api

import (
	"fmt"
	"strings"
)

func (c *CustomQuery) getStreamData(u *UserConfig) string {

	s := `var data = stream
	|from()
	.database('%s')
	.retentionPolicy('%s')
	.measurement('%s')
	.groupBy(%s)`

	whereClause := ""
	if c.QueryDefinition.FilterString != "" {
		whereClause = fmt.Sprintf(".where(lambda: %s)", c.QueryDefinition.FilterString)
	}

	rest := `|window()
	.period(%s)
	.every(%s)
	|%s
	.as('%s')`

	gArray := strings.Split(c.GroupByTag, ",")
	temp := make([]string, len(gArray))
	for i, item := range gArray{
		item = strings.TrimSpace(item)
		temp[i] = fmt.Sprintf("'%s'", item)
	}
	grpString := strings.Join(temp, ",")

	strm := fmt.Sprintf(s, u.DatabaseName, u.RetentionPolicy, c.QueryDefinition.FromString, grpString)
	restStr := fmt.Sprintf(rest, c.Window, c.AlertCheckFrequency,
		c.QueryDefinition.SelectString, c.QueryDefinition.FieldProjectionName)

	finalStr := fmt.Sprintf("%s%s%s", strm, whereClause, restStr)

	return finalStr
}

func (c *CustomQuery) GetStreamAlert(userConfig *UserConfig) string {
	s := `%s
	data
	|alert()
        .id('%s')
        .message(%s)
        %s
        .stateChangesOnly()

        // Alert handlers
        .pagerDuty().serviceKey('%s')
        .log('/tmp/alerts.log') // write to log file
        .slack().channel('%s')`

	return fmt.Sprintf(s, c.getStreamData(userConfig), c.GetAlertId(userConfig), c.getMessage(),
		c.getThresholdStrings(), userConfig.PagerDutyServiceKey, userConfig.SlackChannel)
}