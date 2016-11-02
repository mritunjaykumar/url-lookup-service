package api

import (
	"time"
	"net/http"
	"fmt"
	"strings"
	"encoding/json"
	"github.com/mritunjaykumar/alertTester/metrics"
	"io/ioutil"
	"os/exec"
)

var metricsClient *metrics.StatsdClient

func getMetricsClient(metricsPort int) *metrics.StatsdClient {
	if metricsClient == nil {
		metricsClient, _ = metrics.NewStatsdClient(metricsPort)
	}
	return metricsClient
}

func Deploy(rw http.ResponseWriter, req *http.Request) {
	// Capture application metrics
	metricsTags := []string{
		"url:/deploy",
		fmt.Sprintf("method:%s",req.Method),
	}
	c := getMetricsClient(apiConfig.UDPListenerPortForMetrics)
	go c.IncrementCounter(metricsTags)

	start := time.Now()

	u, customSlices, err := getCustomQueries(req)
	if err != nil {
		http.Error(rw,
			fmt.Sprintf("%s", err.Error()),
			http.StatusBadRequest)
	}

	outputFileNames := make([]string, len(customSlices))
	alertNames := make([]string, len(customSlices))
	alertEnableMap := map[string]string{}

	createTickScriptFiles(customSlices, u, outputFileNames, alertNames, alertEnableMap)

	str := deployAlerts(customSlices, u, alertNames, outputFileNames)

	fmt.Fprintf(rw, "%s", str)

	go c.SendElapsedTime(start, metricsTags)
}

func deployAlerts(customSlice []CustomQuery, u *UserConfig,
alertNames []string, outputFileNames []string) string {
	kapacitorUrl := ""
	// TODO: Hard-coded to avoid any accidental push to Prod
	if strings.Contains(u.KapacitorUrl, "monitoring.prod.aws") || u.KapacitorUrl == "" {
		kapacitorUrl = "http://metrics-alerts.monitoring.nonprod.aws.cloud.nordstrom.net:9092"
	} else {
		kapacitorUrl = u.KapacitorUrl
	}

	resultStr := make([]string, len(customSlice))

	for i, item := range customSlice {
		str := execCommands(kapacitorUrl, alertNames[i], u.DatabaseName, u.RetentionPolicy,
			outputFileNames[i], item.EnableAlertYesNo, u.EnableAllAlertsYesNo)

		resultStr[i] = str
	}

	return strings.Join(resultStr, "\n")
}

func createTickScriptFiles(customSlice []CustomQuery, u *UserConfig,
outputFileNames []string, alertNames []string, alertEnableMap map[string]string) {
	i := 0
	for _, element := range customSlice {
		alertName := element.GetAlertId(u)
		outputFileName := fmt.Sprintf("tickScripts/%s.tick", alertName)

		err := ioutil.WriteFile(outputFileName, []byte(element.GetStreamAlert(u)), 0644)
		check(err)
		outputFileNames[i] = outputFileName
		alertNames[i] = alertName
		alertEnableMap[outputFileName] = element.EnableAlertYesNo
		i++
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func execCommands(kapacitorUrl string, alertName string, db string, rp string, tickScriptFile string,
queryAlertEnable string, allAlertEnable string) string {
	resultSlice := make([]string, 4)

	kapCli := "kapacitor"

	if KapacitorCliPath != "" {
		kapCli = KapacitorCliPath
	}

	delCmd := exec.Command(kapCli, "-url", kapacitorUrl, "delete", "tasks", alertName)

	defineCmd := exec.Command(kapCli, "-url", kapacitorUrl, "define",
		alertName, "-type", "stream", "-tick", tickScriptFile, "-dbrp", fmt.Sprintf("%s.%s", db, rp))

	enableCmd := exec.Command(kapCli, "-url", kapacitorUrl,
		getAction(queryAlertEnable, allAlertEnable), "tasks", alertName)

	listCmd := exec.Command(kapCli, "-url", kapacitorUrl, "list", "tasks", alertName)

	i := 0

	resultSlice[i] = addToResultString(delCmd)
	i++

	resultSlice[i] = addToResultString(defineCmd)
	i++

	resultSlice[i] = addToResultString(enableCmd)
	i++

	resultSlice[i] = addToResultString(listCmd)
	i++

	str := strings.Join(resultSlice, "\n")

	return str
}


func addToResultString(cmd *exec.Cmd) string {
	i := 0
	resSlice := make([]string, 3)

	resSlice[i] = fmt.Sprintf("==> Executing: %s", strings.Join(cmd.Args, " "))
	i++
	output, err := cmd.CombinedOutput()

	if len(output) > 0 {
		resSlice[i] = fmt.Sprintf("==> Output: %s", string(output))
		i++
	}
	if err != nil {
		resSlice[i] = fmt.Sprintf("==> Error: %s", err.Error())
		i++
	}
	if i == 0 {
		return ""
	} else {
		return strings.Join(resSlice, "\n")
	}
}

func getAction(queryEnable string, allEnable string) string {
	action := ""
	if allEnable == "" {
		if strings.ToLower(queryEnable) == "yes" {
			action = "enable"
		} else {
			action = "disable"
		}
	} else {
		if strings.ToLower(allEnable) == "yes" {
			action = "enable"
		} else {
			action = "disable"
		}
	}
	return action
}

func Generate(rw http.ResponseWriter, req *http.Request) {
	// Capture application metrics
	metricsTags := []string{
		"url:/generate",
		fmt.Sprintf("method:%s",req.Method),
	}
	c := getMetricsClient(apiConfig.UDPListenerPortForMetrics)
	go c.IncrementCounter(metricsTags)

	start := time.Now()

	fmt.Println("In generate...")
	u, customSlices, err := getCustomQueries(req)
	if err != nil {
		http.Error(rw,
			fmt.Sprintf("%s", err.Error()),
			http.StatusBadRequest)
	}

	MAX_NUM_OF_ALERT_DEF := 10
	if len(customSlices) > MAX_NUM_OF_ALERT_DEF {
		http.Error(rw,
			fmt.Sprintf("Expected less than %d alert definition, but I get %d.",
				MAX_NUM_OF_ALERT_DEF, len(customSlices)), http.StatusBadRequest)
		return
	}

	alertDefStrings := getAlertDefinitionStrings(u, customSlices)

	separator := "--------- Alert Definition Separator ---------"
	fmt.Fprintf(rw, "%s", strings.Join(alertDefStrings, fmt.Sprintf("\n%s\n", separator)))

	//rw.WriteJson(alertDefs)

	go c.SendElapsedTime(start, metricsTags)
}

func getAlertDefinitionStrings(u *UserConfig, customSlices []CustomQuery) []string {
	alertDefs := make([]string, len(customSlices))
	for i, item := range customSlices {
		alertName := item.GetAlertId(u)
		outputFileName := fmt.Sprintf("%s.tick", alertName)
		startAlertDefString := "------- START OF ALERT DEFINITION (Copy from following line) -------"
		endAlertDefString := "------- END OF ALERT DEFINITION (Copy till above line) -------"

		alertDefs[i] = fmt.Sprintf("Alert Name : %s\n Tick Filename: %s\n%s\n%s\n%s",
			alertName, outputFileName, startAlertDefString,
			item.GetStreamAlert(u), endAlertDefString)
		//item.GetCustomQuery(u), endAlertDefString)
		i++
	}
	return alertDefs
}

func getCustomQueries(req *http.Request) (*UserConfig, []CustomQuery, error) {
	u, err := getUserConfig(req)
	if err != nil{
		return nil, nil, err
	}
	customSlices := u.CustomQueryAlerts[:]
	return u, customSlices, nil
}

func getUserConfig(req *http.Request) (*UserConfig, error) {
	decoder := json.NewDecoder(req.Body)
	u := new(UserConfig)
	err := decoder.Decode(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}