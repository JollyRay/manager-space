package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var delimiter string = "==================="

var conn net.Conn
var serverReader bufio.Reader
var tocken []byte

func main() {
	fmt.Println("Start manager client")

	handleConnection("127.0.0.1:3333")

	reader := bufio.NewReader(os.Stdin)
	handleLogin(reader)
	handleReader(reader)
	conn.Close()

}

func handleConnection(addres string) {
	addresPort := addres
	c, err := net.Dial("tcp", addresPort)
	if err != nil {
		log.Fatal(err)
	}
	conn = c
	serverReader = *bufio.NewReader(c)
}

func handleLogin(reader *bufio.Reader) {
LOGINLOOP:
	for {
		fmt.Print("Login: ")
		login, _ := reader.ReadString('\n')
		fmt.Print("Password: ")
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
		password, _ := reader.ReadString('\n')
		exec.Command("stty", "-F", "/dev/tty", "echo").Run()
		fmt.Println()

		var request []byte = []byte{LOGIN}
		request = append(request, []byte(login)...)
		request = append(request, []byte(password)...)

		_, err := conn.Write(request)

		if err != nil {
			fmt.Println("TCP connection close")
			log.Fatal(err)
		}

		var buf []byte = make([]byte, 9)

		serverReader.Read(buf)

		switch buf[0] {
		case ACCEPT:
			tocken = buf[1:]
			fmt.Print("Welcome ", login)
			break LOGINLOOP
		case CONFIRMATION: //DELETE
			tocken = buf[1:]
			fmt.Print("Pls wait ", login)
			serverReader.Read(buf)
			fmt.Println(buf)
			break LOGINLOOP
		case PERMISSIONDENIED:
			fmt.Println("You was ban")
		case REFUSAL:
			fmt.Println("Login or password miss")
		}
	}
}

const (
	CLOSE        = iota
	HELP         = iota
	SENDTOSERVER = iota
	NEEDREDO     = iota
)

func handleReader(reader *bufio.Reader) {

WORK:
	for {
		fmt.Print("> ")
		value, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		request, ok := requster(value)
		switch ok {
		case CLOSE:
			break WORK
		case HELP:
			fmt.Println("<id> - это элементы после # рядом с пользователями или машинами.\n" +
				"<tocken> - это число у залогиненых в данный момент пользователей\n" +
				"Команды:\n" +
				"rule - показывает все нынешние натройки\n" +
				"order - устанавливает эти настройки\n" +
				"show car - выводит все машины, что есть в базе с информацией о них\n" +
				"show client - выводит всех пользователях, что есть базе с информацией о них" +
				"show queue - выводит токены пользователей, что сейчас находятся в очереди\n" +
				"show pilote - выводит связь <id> пользователя и машины, которой он сейчас управляет\n" +
				"add car <addres> <port> - добавляет в базу новую машинку\n" +
				"add cleint <login> <password> - добавляет в базу нового пилота, за которого смогут залогиниться\n" +
				"add manager <login> <password> - добавляет в базу нового менеджера, за которого смогут залогиниться\n" +
				"delete car <id> - удаляет из базы машину\n" +
				"delete client <id> - удаляет из базы клиента\n" +
				"prepare car <id> - разрешить пользователю из очереди взять машину под <id>\n" +
				"prepare client <id> - разрешить пользователю залогиниться\n" +
				"ban car <id> - если сейчас машина свободна, то следующие пользователи не смогут её занять\n" +
				"ban client <id> - не даёт пользователю залогиниться\n" +
				"break <tocken> - принудительно разрывает соединение по токену с пользователем")
		case SENDTOSERVER:
			_, err := conn.Write(request)
			if err != nil {
				fmt.Println("TCP connection close")
				log.Fatal(err)
			}
			responser(request[9])
		case NEEDREDO:
			fmt.Println("Error comman. You can use \"help\"")
		}
	}
}

func requster(line string) ([]byte, byte) {
	/* line - command from manager
	args[0] - prefix command rule/exit/show/add/delete/break
	args[1] - addition to the command cleint/pilite/car with show/add or index with delete
	args[2] and args[3] - arguments command. Login and Password with add or index with delete
	P.S. exit: len(args) = 1; show/break: len(args) = 2; delete: len(args) = 3; add: len(args) = 4;
	*/
	args := strings.Fields(line)

	if len(args) == 0 {
		return nil, NEEDREDO
	}

	var request []byte = make([]byte, 1)
	request[0] = COMMAND
	request = append(request, tocken...)

	switch strings.ToLower(args[0]) {
	case "exit":
		return nil, CLOSE
	case "help":
		return nil, HELP
	case "rule":
		request = append(request, SHOWRULE)
		return request, SENDTOSERVER
	case "order":
		request = append(request, SETRULE)
		request = append(request, parseOrder(args[1:])...)
		return request, SENDTOSERVER
	case "show":
		if len(args) == 2 {
			args[1] = strings.ToLower(args[1])
			switch args[1] {
			case "car":
				request = append(request, SHOWCAR)
			case "client", "user":
				request = append(request, SHOWCLIENT)
			case "pilote":
				request = append(request, SHOWPILOTE)
			case "queue":
				request = append(request, SHOWQUEUE)
			default:
				return nil, NEEDREDO
			}
			return request, SENDTOSERVER
		}
	case "add":
		args[1] = strings.ToLower(args[1])
		if len(args) == 4 && args[1] == "car" {
			request = append(request, ADDCAR)
			request = append(request, []byte(args[2])...) //TODO: Maybe conver to []byte
			request = append(request, ':')
			request = append(request, []byte(args[3])...) //TODO: Maybe conver to []byte
			return request, SENDTOSERVER
		}
		if len(args) == 4 && args[1] == "client" {
			request = append(request, ADDCLIENT)
			request = append(request, []byte(args[2])...)
			request = append(request, '\n')
			request = append(request, []byte(args[3])...)
			request = append(request, '\n')
			return request, SENDTOSERVER
		}
		if len(args) == 4 && args[1] == "manager" {
			request = append(request, ADDMANAGER)
			request = append(request, []byte(args[2])...)
			request = append(request, '\n')
			request = append(request, []byte(args[3])...)
			request = append(request, '\n')
			return request, SENDTOSERVER
		}
	case "delete":
		if len(args) == 3 {
			switch args[1] {
			case "client":
				request = append(request, DELETECLIENT)
			case "car":
				request = append(request, DELETECAR)
			}
			number, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return nil, NEEDREDO
			}
			for iter := 3; iter > -1; iter-- {
				request = append(request, byte(number>>(iter*8)))
			}
			return request, SENDTOSERVER
		}
	case "prepare":
		if len(args) == 3 {
			switch args[1] {
			case "client":
				request = append(request, PREPARECLIENT)
			case "car":
				request = append(request, PREPARECAR)
			}
			number, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return nil, NEEDREDO
			}
			for iter := 3; iter > -1; iter-- {
				request = append(request, byte(number>>(iter*8)))
			}
			return request, SENDTOSERVER
		}
	case "ban":
		if len(args) == 3 {
			switch args[1] {
			case "client":
				request = append(request, BANCLIENT)
			case "car":
				request = append(request, BANCAR)
			}
			number, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return nil, NEEDREDO
			}
			for iter := 3; iter > -1; iter-- {
				request = append(request, byte(number>>(iter*8)))
			}
			return request, SENDTOSERVER
		}
	case "break":
		if len(args) == 2 {
			request = append(request, CLOSECONNECTION)
			number, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return nil, NEEDREDO
			}
			for iter := 7; iter > -1; iter-- {
				request = append(request, byte(number>>(iter*8)))
			}
			return request, SENDTOSERVER
		}
	}

	return nil, NEEDREDO
}

func responser(kind byte) {
	var headerResponse []byte = make([]byte, 9)
	n, err := serverReader.Read(headerResponse)
	if n != 9 || err != nil {
		log.Fatalf("Unnkown answer from server %v", err)
	}
	for index, value := range headerResponse[1:] {
		if tocken[index] != value {
			log.Fatal("Error tocken")
		}
	}
	if headerResponse[0] == ACCEPT {
		switch kind {
		case SHOWRULE:
			answer := make([]byte, 3)
			_, err := serverReader.Read(answer)
			fmt.Println(answer)
			if err == nil {
				fmt.Println(delimiter)
				deley := (int32(answer[0]&0x0F) << 6) + (int32(answer[1]&0xFC) >> 2)
				fmt.Printf("Минимальный промежуток между сообщениями %d миллисекунд (timeout/to) \n", deley)
				timeout := (int32(answer[1]&0x03) << 8) + int32(answer[2])
				if timeout < 60 {
					fmt.Printf("Максимальное время бездействия пользователя %d секунд (timedeley/td)\n", timeout)
				} else {
					fmt.Printf("Максимальное время бездействия пользователя %d минут %d секунд (timedeley/td)\n", timeout/60, timeout%60)
				}
				c := '-'
				if answer[0]&USERBAN > 0 {
					c = '+'
				}
				fmt.Printf("[%c] Все созданные пользователи по умолчанию доступны (не в бане) (userban/ub)\n", c)
				c = '-'
				if answer[0]&CARBAN > 0 {
					c = '+'
				}
				fmt.Printf("[%c] Все созданные машины по умолчанию доступны (не в бане) (carban/cb)\n", c)
				c = '-'
				if answer[0]&FORCEDISCONNECTUSER > 0 {
					c = '+'
				}
				fmt.Printf("[%c] Пользователь будет разлогинен сразу после его бана (forcedisconnectuser/fdu)\n", c)
				c = '-'
				if answer[0]&FORCEDISCONNECTCAR > 0 {
					c = '+'
				}
				fmt.Printf("[%c] Машина будет разлогинен сразу после его бана (forcedisconnectcar/fdc)\n", c)
				fmt.Println(delimiter)
			}
		case SHOWPILOTE:
			answer, err := serverReader.ReadString(0)
			if err == nil {
				fmt.Println(delimiter)
				for _, item := range strings.Split(answer, ";") {
					part := strings.Split(item, ",")
					if len(part) == 2 {
						fmt.Printf("User#%s drives the car#%s\n", part[0], part[1])
					}
				}
				fmt.Println(delimiter)
			}
		case SHOWCAR:
			answer, err := serverReader.ReadString(0)
			if err == nil {
				fmt.Println(delimiter)
				for _, item := range strings.Split(answer, ";") {
					part := strings.Split(item, ",")
					if len(part) == 3 {
						fmt.Printf("Car#%s have address:port %s", part[0], part[1])
						if part[2] != "true" {
							fmt.Printf(" \tМашина не готова к эксплуатации")
						}
						fmt.Println()
					} else if len(part) == 4 {
						fmt.Printf("Car#%s have address:port %s, pilot user#%s", part[0], part[1], part[3])
						if part[2] != "true" {
							fmt.Printf(" \tМашина не готова к эксплуатации")
						}
						fmt.Println()
					}
				}
				fmt.Println(delimiter)
			}
		case SHOWCLIENT:
			answer, err := serverReader.ReadString(0)
			if err == nil {
				fmt.Println(delimiter)
				for _, item := range strings.Split(answer, ";") {
					part := strings.Split(item, ",")
					if len(part) == 4 {
						fmt.Printf("User#%s Login: %s Password: %s\t", part[0], part[1], part[2])
					} else if len(part) == 5 {
						fmt.Printf("User#%s Login: %s Password: %s tocken: %s\t", part[0], part[1], part[2], part[4])
					} else if len(part) == 6 {
						fmt.Printf("User#%s Login: %s Password: %s now Car#%s tocken: %s\t", part[0], part[1], part[2], part[4], part[5])
					}
					if len(part) > 3 && len(part) < 7 {
						if part[3] == "true" {
							fmt.Print("Пользователь заблокирован, он не сможет снова войти!")
						}
						fmt.Println()
					}
				}
				fmt.Println(delimiter)
			}
		case SHOWQUEUE:
			fmt.Println(delimiter)
			var index int = 1
			for {
				var byteTocken []byte = make([]byte, 8)
				n, err := serverReader.Read(byteTocken)
				if n != 8 {
					break
				}
				if err != nil {
					break
				}
				fmt.Printf("%d. Tocken: %d\n", index, extractTockenToInt63(byteTocken))
				index++
			}
			fmt.Println(delimiter)
		default:
			fmt.Println("Successfully completed")
		}
	} else if headerResponse[0] == PERMISSIONDENIED {
		fmt.Println("Permission denied")
	} else {
		fmt.Println("Server side error, check the data is correct")
	}
}

func extractTockenToInt63(buf []byte) (tocken int64) {
	for _, value := range buf {
		tocken <<= 8
		tocken |= int64(value)
	}
	return
}

func parseOrder(buf []string) []byte {
	request := make([]byte, 0)
	for iter := 0; iter < len(buf); iter++ {
		parts := strings.Split(buf[iter], "=")
		if len(parts) == 2 {
			switch strings.ToLower(parts[0]) {
			case "userban", "ub":
				if strings.Compare(strings.ToLower(parts[1]), "true") == 0 || strings.Compare(strings.ToLower(parts[1]), "t") == 0 || strings.Compare(strings.ToLower(parts[1]), "1") == 0 {
					newOrder := arrangeOrder(USERBAN_NUM, 1)
					request = append(request, newOrder...)
					continue
				}
				if strings.Compare(strings.ToLower(parts[1]), "false") == 0 || strings.Compare(strings.ToLower(parts[1]), "f") == 0 || strings.Compare(strings.ToLower(parts[1]), "0") == 0 {
					newOrder := arrangeOrder(USERBAN_NUM, 0)
					request = append(request, newOrder...)
				}
			case "carban", "cb":
				if strings.Compare(strings.ToLower(parts[1]), "true") == 0 || strings.Compare(strings.ToLower(parts[1]), "t") == 0 || strings.Compare(strings.ToLower(parts[1]), "1") == 0 {
					newOrder := arrangeOrder(CARBAN_NUM, 1)
					request = append(request, newOrder...)
					continue
				}
				if strings.Compare(strings.ToLower(parts[1]), "false") == 0 || strings.Compare(strings.ToLower(parts[1]), "f") == 0 || strings.Compare(strings.ToLower(parts[1]), "0") == 0 {
					newOrder := arrangeOrder(CARBAN_NUM, 0)
					request = append(request, newOrder...)
				}
			case "forcedisconnectuser", "fdu":
				if strings.Compare(strings.ToLower(parts[1]), "true") == 0 || strings.Compare(strings.ToLower(parts[1]), "t") == 0 || strings.Compare(strings.ToLower(parts[1]), "1") == 0 {
					newOrder := arrangeOrder(FORCEDISCONNECTUSER_NUM, 1)
					request = append(request, newOrder...)
					continue
				}
				if strings.Compare(strings.ToLower(parts[1]), "false") == 0 || strings.Compare(strings.ToLower(parts[1]), "f") == 0 || strings.Compare(strings.ToLower(parts[1]), "0") == 0 {
					newOrder := arrangeOrder(FORCEDISCONNECTUSER_NUM, 0)
					request = append(request, newOrder...)
				}
			case "forcedisconnectcar", "fdc":
				if strings.Compare(strings.ToLower(parts[1]), "true") == 0 || strings.Compare(strings.ToLower(parts[1]), "t") == 0 || strings.Compare(strings.ToLower(parts[1]), "1") == 0 {
					newOrder := arrangeOrder(FORCEDISCONNECTCAR_NUM, 1)
					request = append(request, newOrder...)
					continue
				}
				if strings.Compare(strings.ToLower(parts[1]), "false") == 0 || strings.Compare(strings.ToLower(parts[1]), "f") == 0 || strings.Compare(strings.ToLower(parts[1]), "0") == 0 {
					newOrder := arrangeOrder(FORCEDISCONNECTCAR_NUM, 0)
					request = append(request, newOrder...)
				}
			case "timeout", "to":
				if num, err := strconv.ParseUint(parts[1], 10, 16); err == nil {
					newOrder := arrangeOrder(TIMEOUT_NUM, num)
					request = append(request, newOrder...)
				}
			case "timedeley", "td":
				if num, err := strconv.ParseUint(parts[1], 10, 16); err == nil {
					newOrder := arrangeOrder(TIMEDELEY_NUM, num)
					request = append(request, newOrder...)
				}
			}
		}
		//strconv.ParseUint(
	}
	return request
}

func arrangeOrder(rule byte, value uint64) []byte {
	var buf []byte
	if rule == TIMEDELEY_NUM || rule == TIMEOUT_NUM {
		buf = make([]byte, 2)
		buf[0] = 0x80
		buf[0] |= rule << 2
		buf[0] |= byte(value >> 8)
		buf[1] |= byte(value)
	} else {
		buf = make([]byte, 1)
		buf[0] = 0x80
		buf[0] |= rule << 2
		if value == 0 {
			buf[0] &= 0xFC
		} else {
			buf[0] |= 0x03
		}
	}
	return buf
}
