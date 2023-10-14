package main

import (
	"bufio"
	"fmt"
	"k8s.io/apimachinery/pkg/util/json"
	"scan_project/internal/model"
	"strings"
)

func main() {
	logExample := "{\"app\":\"scheduler_go\",\"app_ver\":\"1.6.0\",\"details\":{\"body\":{\"repeat_data\":[{\"domain\":\"27026000002557\",\"group_id\":3}],\"request_id\":\"3730176\"},\"host\":\"csms-server-svc\",\"method\":\"POST\",\"path\":\"/api/v1/request_id/3730176/repeat\"},\"level\":\"trace\",\"module\":\"build_emm_task\",\"msg\":\"\",\"sch_id\":\"v6_activate_pr1\",\"schedule_run_id\":3730176,\"time\":\"2023-10-04T19:12:12.777Z\"}\n{\"app\":\"scheduler_go\",\"app_ver\":\"1.6.0\",\"level\":\"trace\",\"module\":\"build_emm_task\",\"msg\":\"Response status: 200 OK\",\"sch_id\":\"v6_activate_pr1\",\"schedule_run_id\":3730176,\"time\":\"2023-10-04T19:12:12.784Z\"}\n{\"app\":\"scheduler_go\",\"app_ver\":\"1.6.0\",\"details\":{\"body\":{},\"method\":\"POST\",\"path\":\"/api/v1/request_id/3730176/repeat\",\"status\":200},\"level\":\"trace\",\"module\":\"build_emm_task\",\"msg\":\"response\",\"sch_id\":\"v6_activate_pr1\",\"schedule_run_id\":3730176,\"time\":\"2023-10-04T19:12:12.784Z\"}\n{\"app\":\"scheduler_go\",\"app_ver\":\"1.6.0\",\"args\":[\"28026000002558\",3,\"000000000000000202000000000001010003090308040006034f9606030596\",0,26,\"tricolor\",6],\"cid\":\"a25511c0-d053-4eb9-8460-1e7b1947d5f1\",\"level\":\"error\",\"msg\":\"Query\",\"pid\":2548,\"sql\":\"select * from casapi010200.get_keys($1, $2, $3, $4, $5, $6, $7)\",\"time\":\"2023-10-04T19:12:12.786Z\"}\n{\"app\":\"scheduler_go\",\"app_ver\":\"1.6.0\",\"cid\":\"a25511c0-d053-4eb9-8460-1e7b1947d5f1\",\"level\":\"error\",\"module\":\"build_emm_task\",\"msg\":\"failed to get next tuple: failed to get cas keys: ERROR: Function get_keys. Uid not found: 2558 (SQLSTATE P0001)\",\"sch_id\":\"v6_activate_pr1\",\"schedule_run_id\":3730176,\"time\":\"2023-10-04T19:12:12.786Z\"}\n{\"app\":\"scheduler_go\",\"app_ver\":\"1.6.0\",\"level\":\"debug\",\"module\":\"build_emm_task\",\"msg\":\"Request: POST csms-server-svc/api/v1/request_id/3730176/repeat\",\"sch_id\":\"v6_activate_pr1\",\"schedule_run_id\":3730176,\"time\":\"2023-10-04T19:12:12.786Z\"}"
	r := strings.NewReader(logExample) // r type is io.ReadCloser
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log := &model.CommonServiceLog{}
		err := json.Unmarshal(scanner.Bytes(), log)
		if err != nil {
			panic(err)
		}
		fmt.Println(*log.Level)
	}
}
