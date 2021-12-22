package user

import (
	"fmt"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/thank243/StairUnlocker-Bot/config"
	"github.com/thank243/StairUnlocker-Bot/utils"
	"gopkg.in/yaml.v3"
	"sort"
	"strings"
	"time"
)

func (u *User) URLCheck() {
	var proxiesList []C.Proxy
	u.IsCheck = true
	defer func() { u.IsCheck = false }()
	proxies, unmarshalProxies, err := u.generateProxies(config.BotCfg.ConverterAPI)
	if err != nil {
		_ = u.EditMessage(u.MessageID, err.Error())
		return
	}
	_ = u.EditMessage(u.MessageID, "Checking nodes unlock status...")
	for _, v := range proxies {
		proxiesList = append(proxiesList, v)
	}
	//同时连接数
	connNum := config.BotCfg.MaxConn
	if i := len(proxiesList); i < connNum {
		connNum = i
	}
	start := time.Now()
	netflixList, latency := utils.BatchCheck(proxiesList, connNum)
	sort.Strings(netflixList)
	sort.Strings(latency)
	//proxiesTest(netflixList,u)
	log.Warnln(fmt.Sprintf("Total %d nodes, %d unlock nodes. Elapsed time: %s", len(proxiesList), len(netflixList), time.Since(start).Round(time.Millisecond)))
	report := fmt.Sprintf("Total %d nodes, %d unlock nodes.\nElapsed time: %s", len(proxiesList), len(netflixList), time.Since(start).Round(time.Millisecond))
	telegramReport := fmt.Sprintf("%s\nTimestamp: %s\n%s\n%s", report, time.Now().Round(time.Millisecond), strings.Repeat("-", 35), strings.Join(latency, "\n"))
	u.Data.CheckInfo = telegramReport
	// update message, not send a new message
	_ = u.EditMessage(u.MessageID, telegramReport)
	// send proxies.yaml
	if len(netflixList) > 0 {
		marshal, _ := yaml.Marshal(NetflixFilter(netflixList, unmarshalProxies))
		_, err = u.Bot.Send(tgbotapi.NewDocument(u.ID, tgbotapi.FileBytes{
			Name:  "proxies.yaml",
			Bytes: marshal,
		}))
	}
	if err != nil {
		log.Errorln(err.Error())
	}
}
