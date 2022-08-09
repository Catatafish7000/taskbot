package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type User struct {
	Name string
}

const (
	BotToken   = "5580133190:AAFxCSFA2BCGBFVb-7VudnyuBL6vmZi8QKM"
	WebhookURL = "https://b4fd-85-89-126-36.eu.ngrok.io"
)

type Task struct {
	Name   string
	Worker int64
	Owner  int64
}

type BotData struct {
	Users map[int64]User
	Tasks []Task
	Bot   *tgbotapi.BotAPI
}

func (data *BotData) Start(upd tgbotapi.Update) {
	msg := "/tasks - показать все задачи\n" +
		"/new - создать новую задачу\n" +
		"/my - показать свои задачи\n" +
		"/owner - показать собственноручно созданные задачи\n" +
		"/assign_x - взять себе задачу под номером x\n" +
		"/unassign_x - снять с себя задачу подномером x\n" +
		"/resolve_x - обозначить задачу под номером x как выполненную\n" +
		"/info - вернуться к этому экрану"
	data.SendMsg(upd, msg)
}
func (data *BotData) SendMsg(upd tgbotapi.Update, msg string) {
	ans := tgbotapi.NewMessage(upd.Message.Chat.ID, msg)
	_, err := data.Bot.Send(ans)
	if err != nil {
		return
	}
}

func (data *BotData) ShowTasks(upd tgbotapi.Update) {
	if len(data.Tasks) == 0 {
		data.SendMsg(upd, "Нет задач")
		return
	}
	ans := ""
	for i, task := range data.Tasks {
		ans += strconv.Itoa(i) + ") " + task.Name
		if task.Worker != 0 {
			ans += " выполняет пользователь @" + data.Users[task.Worker].Name + "\n\n"
		} else {
			ans += "\n/assign_" + strconv.Itoa(i) + "\n\n"
		}
	}
	data.SendMsg(upd, ans)
}

func (data *BotData) CreateTask(upd tgbotapi.Update) {
	task := upd.Message.CommandArguments()
	if task == "" {
		msg := "Нужно написать название задания"
		data.SendMsg(upd, msg)
		return
	}
	id := len(data.Tasks)
	data.Tasks = append(data.Tasks, Task{
		Name:   task,
		Worker: 0,
		Owner:  upd.Message.From.ID,
	})
	msg := "Задача \"" + task + "\" создана с идентификатором " + strconv.Itoa(id)
	data.SendMsg(upd, msg)
}
func (data *BotData) Assign(upd tgbotapi.Update, num string) {
	taskID, err := strconv.Atoi(num)
	if err != nil {
		msg := "Нужно ввести номер задания"
		data.SendMsg(upd, msg)
		return
	}
	data.Tasks[taskID].Worker = upd.Message.From.ID
	msg := "Вы взяли задание номер " + num + " \"" + data.Tasks[taskID].Name
	data.SendMsg(upd, msg)
}

func (data *BotData) Unassign(upd tgbotapi.Update, num string) {
	taskID, err := strconv.Atoi(num)
	if err != nil {
		msg := "Нужно ввести номер задания"
		data.SendMsg(upd, msg)
		return
	}
	if data.Tasks[taskID].Worker != upd.Message.From.ID {
		msg := "Чужая задача"
		data.SendMsg(upd, msg)
		return
	}
	data.Tasks[taskID].Worker = 0
	msg := "Вы сняли с себя задачу номер " + num + " \"" + data.Tasks[taskID].Name + "\""
	data.SendMsg(upd, msg)
}

func (data *BotData) Resolve(upd tgbotapi.Update, num string) {
	taskID, err := strconv.Atoi(num)
	if err != nil {
		msg := "Нужно ввести номер задания"
		data.SendMsg(upd, msg)
		return
	}
	if data.Tasks[taskID].Worker != upd.Message.From.ID {
		msg := "Чужая задача"
		data.SendMsg(upd, msg)
		return
	}
	name := data.Tasks[taskID].Name
	if taskID != len(data.Tasks)-1 {
		data.Tasks = append(data.Tasks[:taskID], data.Tasks[taskID+1:]...)
	} else {
		data.Tasks = data.Tasks[:taskID]
	}
	//copy(data.Tasks[taskID:], data.Tasks[taskID+1:])
	//data.Tasks = data.Tasks[:len(data.Tasks)-1]
	msg := "Вы решили задачу \"" + name + "\""
	data.SendMsg(upd, msg)
}

func (data *BotData) My(upd tgbotapi.Update) {
	ans := ""
	for i, task := range data.Tasks {
		if task.Worker == upd.Message.From.ID {
			ans += strconv.Itoa(i) + ") \"" + task.Name + "\"\n" +
				"/unassign_" + strconv.Itoa(i) + "  /resolve_" + strconv.Itoa(i) + "\n\n"
		}
	}
	if ans == "" {
		data.SendMsg(upd, "У вас нет задач")
		return
	}
	data.SendMsg(upd, ans)
}

func (data *BotData) Owner(upd tgbotapi.Update) {
	ans := ""
	for i, task := range data.Tasks {
		if task.Owner == upd.Message.From.ID {
			ans += strconv.Itoa(i) + ") \"" + task.Name + "\"\n" +
				"/assign_" + strconv.Itoa(i) + "\n\n"
		}
	}
	if ans == "" {
		data.SendMsg(upd, "У вас нет задач")
		return
	}
	data.SendMsg(upd, ans)
}

func (data *BotData) Commander(upd tgbotapi.Update, comm string) {
	switch {
	case comm == "info":
		data.Start(upd)
	case comm == "my":
		data.My(upd)
	case comm == "owner":
		data.Owner(upd)
	case comm == "tasks":
		data.ShowTasks(upd)
	case comm == "new":
		data.CreateTask(upd)
	case strings.Contains(comm, "_"):
		compositeComm := strings.Split(comm, "_")
		switch compositeComm[0] {
		case "unassign":
			data.Unassign(upd, compositeComm[1])
		case "assign":
			data.Assign(upd, compositeComm[1])
		case "resolve":
			data.Resolve(upd, compositeComm[1])
		default:
			data.SendMsg(upd, "Неизвестная команда\nНажмите /info")
			return
		}
	default:
		data.SendMsg(upd, "Неизвестная команда\nНажмите /info")
		return
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatalf("New bot API failed: %s", err)
	}
	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)
	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		log.Fatalf("NewWebhook failed: %s", err)
	}
	_, err = bot.Request(wh)
	if err != nil {
		log.Fatalf("SetWebhook failed: %s", err)
	}
	updates := bot.ListenForWebhook("/")
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("all is working"))
		if err != nil {
			return
		}
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func() {
		log.Fatalln("http err:", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("start listen :" + port)
	users := make(map[int64]User)
	tasks := make([]Task, 0)
	var botInfo = BotData{
		Tasks: tasks,
		Users: users,
		Bot:   bot,
	}
	for update := range updates {
		log.Printf("upd: %#v\n", update)
		if update.Message == nil {
			log.Println("Change of message")
			continue
		}
		if update.Message.From == nil {
			log.Println("nil User")
			continue
		}
		_, ok := botInfo.Users[update.Message.From.ID]
		if !ok {
			botInfo.Users[update.Message.From.ID] = User{
				Name: update.Message.From.UserName,
			}
		}
		command := update.Message.Command()
		if command == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "That's not a command")
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Sending failed: %s", errSend)
			}
			continue
		}
		botInfo.Commander(update, command)
	}
}
