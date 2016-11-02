package api

import (
	"fmt"
	"strings"
)

type (
	UserConfig struct {
		KapacitorUrl         string `json:"kapacitorUrl,omitempty"`
		EnableAllAlertsYesNo string
		DatabaseName         string
		RetentionPolicy      string
		PagerDutyServiceKey  string
		SlackChannel         string
		CustomQueryAlerts    []CustomQuery
	}

	CustomQuery struct {
		QueryDefinition     Query
		EnableAlertYesNo    string
		GroupByTag          string
		CustomMessage       string `json:"customMessage,omitempty"`
		Window              string
		AlertCheckFrequency string
		InfoThreshold       Threshold
		WarnThreshold       Threshold
		CriticalThreshold   Threshold
	}

	Query struct {
		SelectString        string
		FieldProjectionName string
		FromString          string
		FilterString        string
	}

	Threshold struct {
		ComparisonOperator string
		ThresholdValue     string
	}

	StreamParameters struct {
		info string
		warn string
		crit string
		window string
		alertCheckFrequency string
	}
)

func (q *Query) getQueryDef(db string, rp string) string {
	if q.FilterString == "" {
		return fmt.Sprintf("%s as %s from \"%s\".\"%s\".\"%s\"", q.SelectString,
			q.FieldProjectionName, db, rp, q.FromString)
	} else {
		return fmt.Sprintf("%s as %s from \"%s\".\"%s\".\"%s\" where %s", q.SelectString,
			q.FieldProjectionName, db, rp, q.FromString, q.FilterString)
	}
}

func (t *Threshold) getThreshold(field string) string{
	return fmt.Sprintf("lambda: \"%s\" %s %s", field, t.ComparisonOperator, t.ThresholdValue)
}

func (c *CustomQuery) getMessage() string {
	s := `'Alert ID {{ .ID }}: %s of %s is at {{ .Level }} level.;
            {{ .Group }}
            Current value for %s: {{ index .Fields "%s" }}, warning threshold: %s, critical threshold: %s. %s'`

	customMsg := ""
	if c.CustomMessage != "" {
		customMsg = c.CustomMessage
	}
	return fmt.Sprintf(s, c.QueryDefinition.FieldProjectionName,
		c.QueryDefinition.FromString, c.QueryDefinition.FieldProjectionName,
		c.QueryDefinition.FieldProjectionName,
		c.WarnThreshold.ThresholdValue, c.CriticalThreshold.ThresholdValue, customMsg)
}

func (c *CustomQuery) GetAlertId(u *UserConfig) string {
	groupByString := ""
	if(c.GroupByTag != ""){
		sArray := strings.Split(c.GroupByTag, ",")
		temp := make([]string, len(sArray))
		for i, item := range sArray{
			temp[i] = strings.TrimSpace(item)
		}
		groupByString = strings.Join(temp, "_")
	}
	return fmt.Sprintf("%s_%s_%s_%s_%s", u.DatabaseName, u.RetentionPolicy,
		c.QueryDefinition.FromString, c.QueryDefinition.FieldProjectionName, groupByString)
}

func (c *CustomQuery) getThresholdStrings() string {
	tString := ""

	if(c.InfoThreshold.ThresholdValue != ""){
		if(tString == "") {
			tString = fmt.Sprintf(".info(%s)",
				c.InfoThreshold.getThreshold(c.QueryDefinition.FieldProjectionName))
		} else {
			tString = fmt.Sprintf("%s.info(%s)", tString,
				c.InfoThreshold.getThreshold(c.QueryDefinition.FieldProjectionName))
		}
	}
	if(c.WarnThreshold.ThresholdValue != ""){
		if(tString == "") {
			tString = fmt.Sprintf(".warn(%s)",
				c.WarnThreshold.getThreshold(c.QueryDefinition.FieldProjectionName))
		} else {
			tString = fmt.Sprintf("%s.warn(%s)", tString,
				c.WarnThreshold.getThreshold(c.QueryDefinition.FieldProjectionName))
		}
	}
	if(c.CriticalThreshold.ThresholdValue != ""){
		if(tString == "") {
			tString = fmt.Sprintf(".crit(%s)",
				c.CriticalThreshold.getThreshold(c.QueryDefinition.FieldProjectionName))
		} else {
			tString = fmt.Sprintf("%s.crit(%s)", tString,
				c.CriticalThreshold.getThreshold(c.QueryDefinition.FieldProjectionName))
		}
	}
	return tString
}

func (c *CustomQuery) GetCustomQuery(userConfig *UserConfig) string {
	gArray := strings.Split(c.GroupByTag, ",")
	temp := make([]string, len(gArray))
	for i, item := range gArray{
		item = strings.TrimSpace(item)
		temp[i] = fmt.Sprintf("'%s'", item)
	}
	grpString := strings.Join(temp, ",")
	s := `var data = batch
    |query('''
        %s
    ''')
    .groupBy(%s)
    .period(%s)
    .every(%s)
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

	return fmt.Sprintf(s, c.QueryDefinition.getQueryDef(userConfig.DatabaseName, userConfig.RetentionPolicy),
		grpString, c.Window, c.AlertCheckFrequency, c.GetAlertId(userConfig),
		c.getMessage(), c.getThresholdStrings(), userConfig.PagerDutyServiceKey, userConfig.SlackChannel)
}







