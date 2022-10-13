package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
	tgBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/thank243/StairUnlocker-Bot/model"
	"github.com/thank243/StairUnlocker-Bot/utils"
)

func statistic(streamMediaList *[]model.StreamData) map[string]int {
	statMap := make(map[string]int)
	for i := range *streamMediaList {
		statMap[(*streamMediaList)[i].Name]++
		if !(*streamMediaList)[i].Unlock {
			statMap[(*streamMediaList)[i].Name]--
		}
	}
	return statMap
}

func (u *User) streamMedia(subUrl string) error {
	u.isCheck.Store(true)

	var proxiesList []C.Proxy
	checkFlag := make(chan bool)
	defer func() {
		u.data.lastCheck.Store(time.Now().Unix())
		u.isCheck.Store(false)
		close(checkFlag)
	}()

	msgInst, _ := u.SendMessage("Converting from API server.")
	proxies, err := u.buildProxies(subUrl)
	if err != nil {
		u.EditMessage(msgInst.MessageID, err.Error())
		return err
	}
	if subUrl != "" {
		u.data.subURL.Store(subUrl)
	}

	// animation while waiting test.
	go u.loading("Checking nodes unlock status", checkFlag, msgInst.MessageID)


	for _, v := range proxies {
		proxiesList = append(proxiesList, v)
	}
	// Must have valid node.
	if len(proxiesList) > 0 {
		start := time.Now()
		unlockList := utils.BatchCheck(proxiesList, model.BotCfg.MaxConn)
		checkFlag <- true
		report := fmt.Sprintf("Total %d nodes, Duration: %s", len(proxiesList), time.Since(start).Round(time.Millisecond))

		var nameList []string
		statisticMap := statistic(&unlockList)
		for k := range statisticMap {
			nameList = append(nameList, k)
		}
		sort.Strings(nameList)
		var finalStr string
		for i := range nameList {
			finalStr += fmt.Sprintf("%s: %d\n", nameList[i], statisticMap[nameList[i]])
		}
		telegramReport := fmt.Sprintf("StairUnlocker Bot %s Bulletin:\n%s\n%sTimestamp: %s\n%s", C.Version, report, finalStr, time.Now().UTC().Format(time.RFC3339), strings.Repeat("-", 25))
		log.Warnln("[ID: %d] %s", u.ID, report)
		u.EditMessage(msgInst.MessageID, "Uploading PNG file...")

		buffer, err := utils.GeneratePNG(unlockList, nameList)
		if err != nil {
			return err
		}
		// send result image
		wrapPNG := tgBot.NewDocument(u.ID, tgBot.FileBytes{
			Name:  fmt.Sprintf("stairunlocker_bot_result_%d.png", time.Now().Unix()),
			Bytes: buffer.Bytes(),
		})
		wrapPNG.Caption = fmt.Sprintf("%s\n@stairunlock_test_bot\nProject: https://git.io/Jyl5l", telegramReport)
		// save test results.
		u.data.checkedInfo.Store(wrapPNG.Caption)

		u.s.Bot.Send(wrapPNG)
		u.DeleteMessage(msgInst.MessageID)
	}
	return nil
}
