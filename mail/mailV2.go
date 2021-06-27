package mail

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/sloonz/go-iconv"
	"github.com/sloonz/go-qprintable"
	"github.com/ziutek/mymysql/autorc"
	_ "github.com/ziutek/mymysql/godrv"
)

var Config = map[string]string{
	"MAX_SMTP_CLIENTS":      "10000",
	"SMTP_MAX_SIZE":         "131072",
	"SMTP_HOST_NAME":        "smtp.exmail.qq.com:465",
	"SMTP_VERBOSE":          "Y",
	"SMTP_TIMEOUT":          "100",
	"MYSQL_HOST":            "127.0.0.1:3306",
	"MYSQL_USER":            "mail_it",
	"MYSQL_PASS":            "12345678",
	"MYSQL_DB":              "mail_it",
	"MAIL_TABLE":            "new_mail",
	"SMTP_USER":             "wujiabang@actionsky.com",
	"STMP_LISTEN_INTERFACE": "1.0.0.0:25",
	"STMP_LOG_FILE":         "mail_it.log",
	"SMTP_GID":              "",
	"SMTP_UID":              "",
	"SMTP_PUB_KEY":          "/etc/ssl/certs/ssl-cert-snakeoil.pem",
	"SMTP_PRV_KEY":          "/etc/ssl/private/ssl-cert-snakeoil.key",
	"ALLOWED_HOSTS":         "",
	"PRIMARY_MAIL_HOSTS":    "smtp.exmail.qq.com",
	"CONN_BACKLOG":          "100",
	"MAX_CLIENTS":           "500",
	"SGID":                  "1008",
	"SUID":                  "1008",
}

type Client struct {
	state       int
	helo        string
	mail_from   string
	rcpt_to     string
	read_buffer string
	response    string
	address     string
	data        string
	subject     string
	hash        string
	time        int64
	tls_on      bool
	socket      net.Conn
	bufin       *bufio.Reader
	bufout      *bufio.Writer
	kill_time   int64
	errors      int
	clientID    int64
	savedNotify chan int
}

type redisClient struct {
	count int
	conn  redis.Conn
	time  int
}

var TLSconfig *tls.Config
var clientChan chan *Client

var sem chan int              // currently active clients
var SaveMailChan chan *Client //workers for saving mail

// hosts allowed in the 'to' address
var allowedHosts = make(map[string]bool, 15)

func configure() {
	var configFile, verbose, iface string
	flag.StringVar(&configFile, "config", "exmail.conf", "Path to the exmail configuration file")
	flag.StringVar(&verbose, "v", "n", "Verbose, [y | n ]")
	flag.StringVar(&iface, "if", "", "Interface and port to listen on,eg. 127.0.0.1:2525")
	flag.Parse()
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Could not read config file")
		panic(err)
	}
	var myConfig map[string]string
	err = json.Unmarshal(b, &myConfig)
	if err != nil {
		fmt.Println("Could not parse config file")
		panic(err)
	}
	fmt.Printf("%#v\n", myConfig)
	for k, v := range myConfig {
		Config[k] = v
	}
	Config["SMTP_VERBOSE"] = strings.ToUpper(verbose)
	if len(iface) > 0 {
		Config["STMP_LISTEN_INTERFACE"] = iface
	}
	if arr := strings.Split(Config["ALLOWED_HOSTS"], ","); len(arr) > 0 {
		for i := 0; i < len(arr); i++ {
			allowedHosts[arr[i]] = true
		}
	}
	var n int
	var n_err error
	if n, n_err = strconv.Atoi(Config["MAX_CLIENTS"]); n_err != nil {
		n = 50
	}

	//currently active client list
	sem = make(chan int, n)

	// database writing workers
	SaveMailChan = make(chan *Client, 4)
	return
}

func logln(level int, s string) {
	if level == 2 {
		log.Fatalf(s)
	}
	if Config["SMTP_VERBOSE"] == "Y" {
		fmt.Println(s)
	}
}

func MailV2() {
	configure()
	logln(1, "Loading priv:"+Config["SMTP_PRV_KEY"]+" and pub:"+Config["SMTP_PUB_KEY"])
	// cert, err := tls.LoadX509KeyPair(Config["SMTP_PUB_KEY"], Config["SMTP_PRV_KEY"])
	// if err != nil {
	// 	logln(2, fmt.Sprintf("Cannot listen on port: %s", err))
	// }
	// TLSconfig = &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.VerifyClientCertIfGiven, ServerName: Config["SMTP_HOST_NAME"]}
	// TLSconfig.Rand = rand.Reader
	servername, _, _ := net.SplitHostPort(Config["SMTP_HOST_NAME"])
	TLSconfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         servername,
	}
	listener, err := net.Listen("tcp", Config["STMP_LISTEN_INTERFACE"])
	if err != nil {
		logln(2, fmt.Sprintf("Cannot listen on port, %s", err))
	}
	gid, _ := strconv.ParseInt(Config["SGID"], 10, 32)
	uid, _ := strconv.ParseInt(Config["SUID"], 10, 32)
	syscall.Setgid(int(gid))
	syscall.Setuid(int(uid))
	logln(1, fmt.Sprintf("server listening on"+Config["STMP_LISTEN_INTERFACE"]))
	go Serve(clientChan)
	go saveMail()
	clientID := int64(1)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logln(1, fmt.Sprintf("Accept error: %s", err))
			break
		}
		logln(1, fmt.Sprintf("server: accept from %s", conn.RemoteAddr()))
		// place a new client on the channel
		clientChan <- &Client{
			socket:      conn,
			address:     conn.RemoteAddr().String(),
			time:        time.Now().Unix(),
			bufin:       bufio.NewReader(conn),
			bufout:      bufio.NewWriter(conn),
			clientID:    clientID,
			savedNotify: make(chan int),
		}
		clientID++
	}
}

func Serve(clientChan chan *Client) {
	for {
		c := <-clientChan
		sem <- 1
		go handleClient(c)
		logln(1, fmt.Sprintf("There are now"+strconv.Itoa(runtime.NumGoroutine())+"goroutines"))
	}
}

func readSmtp(client *Client) (input string, err error) {
	var reply string
	// Command state terminator by default
	suffix := "\r\n"
	if client.state == 2 {
		// DATA state
		suffix = "\r\n.\r\n"
	}
	for err == nil {
		client.socket.SetDeadline(time.Now().Add(100 * time.Second))
		reply, err = client.bufin.ReadString('\n')
		if reply != "" {
			input = input + reply
			if client.state == 2 {
				scanSubject(client, reply)
			}
		}
		if err != nil {
			break
		}
		if strings.HasPrefix(input, suffix) {
			break
		}
	}
	return input, err
}

func scanSubject(client *Client, reply string) {
	if client.subject == "" && (len(reply) > 8) {
		test := strings.ToUpper(reply[0:9])
		if i := strings.Index(test, "SUBJECT: "); i == 0 {
			// first line with \r\n
			client.subject = reply[9:]
		}
	} else if strings.HasSuffix(client.subject, "\r\n") {
		// chop  off the \r\n
		client.subject = client.subject[0 : len(client.subject)-2]
		if strings.HasPrefix(reply, " ") || (strings.HasPrefix(reply, "\t")) {
			// subject is multi-line
			client.subject = client.subject + reply[1:]
		}
	}
}

func responseWrite(client *Client) (err error) {
	var size int
	client.socket.SetDeadline(time.Now().Add(100 * time.Second))
	size, err = client.bufout.WriteString(client.response)
	client.bufout.Flush()
	client.response = client.response[size:]
	return err
}

func responseAdd(client *Client, line string) {
	client.response = line + "\r\n"
}

func responseClear(client *Client) {
	client.response = ""
}

func killClient(client *Client) {
	client.kill_time = time.Now().Unix()
}

func closeClient(client *Client) {
	client.socket.Close()
	<-sem
}

func handleClient(client *Client) {
	var input_hist string
	defer closeClient(client)
	greeting := "220" + Config["SMTP_HOST_NAME"] + "SMTP HELLO #" +
		strconv.FormatInt(client.clientID, 10) + " (" + strconv.Itoa(len(sem)) + ") " + time.Now().Format(time.RFC1123Z)
	advertiseTls := "250-STARTTLS\r\n"
	for i := 0; i < 10; i++ {
		switch client.state {
		case 0:
			responseAdd(client, greeting)
			client.state = 1
		case 1:
			input, err := readSmtp(client)
			if err != nil {
				if err == io.EOF {
					// client closed the connection already
					return
				}
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					// too slow, timeout
					return
				}
			}
			input = strings.Trim(input, "\n\r")
			input_hist = input_hist + input + "\n"
			cmd := strings.ToUpper(input)
			switch {
			case strings.Index(cmd, "HELO") == 0:
				if len(input) > 5 {
					client.helo = input[5:]
				}
				responseAdd(client, "250 "+Config["SMTP_HOST_NAME"]+" Hello"+client.helo+"["+client.address+"]"+
					"\r\n"+"250-SIZE "+Config["SMTP_MAX_SIZE"]+"\r\n"+advertiseTls+"250 HELP")

			case strings.Index(cmd, "EHLO") == 0:
				if len(input) > 5 {
					client.helo = input[5:]
				}
				if client.tls_on {
					advertiseTls = ""
				}
				responseAdd(client, "250-"+Config["SMTP_HOST_NAME"]+" Hello "+client.helo+
					"["+client.address+"]"+"\r\n"+"250-SIZE "+Config["SMTP_MAX_SIZE"]+"\r\n"+
					advertiseTls+"250 HELP")
			case strings.Index(cmd, "MAIL FROM:") == 0:
				if len(input) > 10 {
					client.mail_from = input[10:]
				}
				responseAdd(client, "250 Ok")
			case strings.Index(cmd, "RCPT TO:") == 0:
				if len(input) > 8 {
					client.rcpt_to = input[8:]
				}
				responseAdd(client, "250 Accepted")
			case strings.Index(cmd, "NOOP") == 0:
				responseAdd(client, "250 OK")
			case strings.Index(cmd, "REST") == 0:
				client.mail_from = ""
				client.rcpt_to = ""
				responseAdd(client, "250 OK")
			case strings.Index(cmd, "DATA") == 0:
				responseAdd(client, "354 Enter message, ending with \".\" on a line by itself")
				client.state = 2
			case (strings.Index(cmd, "STARTTLS") == 0) && !client.tls_on:
				responseAdd(client, "220 Ready to start TLS")
				//go to start TLS state
				client.state = 3
			case strings.Index(cmd, "QUIT") == 0:
				responseAdd(client, "221 Bye")
				killClient(client)
			default:
				responseAdd(client, fmt.Sprintf("500 unrecognized command %v", err))
				client.errors++
				if client.errors > 3 {
					responseAdd(client, fmt.Sprintf("500 Too many unrecognized commands %v", err))
					killClient(client)
				}
			}
		case 2:
			var err error
			client.data, err = readSmtp(client)
			if err == nil {
				//to do: timeout when adding to SaveMailChan
				//place on the channel so that one of the save mail workers can pick it up
				SaveMailChan <- client
				//wait for the save to complete
				status := <-client.savedNotify
				if status == 1 {
					responseAdd(client, "250 OK : queued as "+client.hash)
				} else {
					responseAdd(client, "554 Error: transaction failed")
				}
			}
			client.state = 1
		case 3:
			//upgrade to TLS
			var tlsConn *tls.Conn
			tlsConn = tls.Server(client.socket, TLSconfig)
			tlsConn.Handshake()
			client.socket = net.Conn(tlsConn)
			client.bufin = bufio.NewReader(client.socket)
			client.bufout = bufio.NewWriter(client.socket)
			client.state = 1
			client.tls_on = true
		}
		//Send a response back to the client
		err := responseWrite(client)
		if err != nil {
			if err == io.EOF {
				// client closed the connection already
				return
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				// too slow, timeout
				return
			}
		}
		if client.kill_time > 1 {
			return
		}
	}
}

func saveMail() {
	var to string
	var err error
	var body string
	var redis_err error
	var length int
	redis := &redisClient{}
	db := autorc.New("tcp", "", Config["MYSQL_HOST"], Config["MYSQL_USER"], Config["MYSQL_PASS"], Config["MYSQL_DB"])
	db.Register("set names utf8")
	sql := "INSERT INTO " + Config["M_MAIL_TABLE"] + " "
	sql += "(`date`,`to`,`from`,`subejct`,`body`,`charset`,`mail`,`spam_score`,`hash`," +
		"`content_type`,`recepient`,`has_attach`,`ip_addr`)"
	sql += " values (NOW(), ?, ?, ?, ? ,'UTF-8' , ?, 0, ?, '', ?, 0, ?)"
	ins, sql_err := db.Prepare(sql)
	if sql_err != nil {
		logln(2, fmt.Sprintf("Sql statement incorrect: %s", sql_err))
	}
	sql = "UPDATE gm2_setting SET `setting_value` = `setting_value` + 1 WHERE `setting_name` = 'received_emails' LIMIT 1"
	incr, sql_err := db.Prepare(sql)
	if sql_err != nil {
		logln(2, fmt.Sprintf("Sql statement incorrect: %s", sql_err))
	}

	for {
		client := <-SaveMailChan
		if user, _, addr_err := validateEmailData(client); addr_err != nil {
			logln(1, fmt.Sprintln("mail_from didnt validate: %v", addr_err)+"client.mail_from:"+client.mail_from)
			client.savedNotify <- 1
			continue
		} else {
			to = user + "@" + Config["M_PRIMARY_MAIL_HOST"]
		}
		length = len(client.data)
		client.subject = mimeHeaderDecode(client.subject)
		client.hash = md5hex(to + client.mail_from +
			client.subject + strconv.FormatInt(time.Now().UnixNano(), 10))
		// Add extra headers
		add_header := ""
		add_header += "Delivered-To: " + to + "\r\n"
		add_header += "Received: from " + client.helo + " (" + client.helo + " [" +
			client.address + "])\r\n"
		add_header += " by " + Config["SMTP_HOST_NAME"] + " with SMTP id " + client.hash +
			"@" + Config["SMTP_HOST_NAME"] + ";\r\n"
		add_header += "   " + time.Now().Format(time.RFC1123Z) + "\r\n"

		// compare to save space
		client.data = compress(add_header + client.data)
		body = "gzencode"
		redis_err = redis.redisConnection()
		if redis_err == nil {
			_, do_err := redis.conn.Do("SETEX", client.hash, 3600, client.data)
			if do_err == nil {
				client.data = ""
				body = "redis"
			}
		} else {
			fmt.Println("redis err", redis_err)
		}
		ins.Bind(
			to,
			client.mail_from,
			client.subject,
			body,
			client.data,
			client.hash,
			to,
			client.address)
		//save, discard result
		_, _, err = ins.Exec()
		if err != nil {
			logln(1, fmt.Sprintf("Database error, %v %v", err))
			client.savedNotify <- -1
		} else {
			logln(1, " Email saved "+client.hash+" len:"+strconv.Itoa(length))
			_, _, err = incr.Exec()
			if err != nil {
				fmt.Println(err)
			}
			client.savedNotify <- 1
		}
	}

}

func validateEmailData(client *Client) (user string, host string, addr_err error) {
	if user, host, addr_err = extractEmail(client.mail_from); addr_err != nil {
		return user, host, addr_err
	}
	client.mail_from = user + "@" + host
	if user, host, addr_err = extractEmail(client.rcpt_to); addr_err != nil {
		return user, host, addr_err
	}
	client.rcpt_to = user + "@" + host
	// check if on allowed hosts
	if allowed := allowedHosts[host]; !allowed {
		return user, host, errors.New("invalid host:" + host)
	}
	return user, host, addr_err
}

func extractEmail(str string) (name string, host string, err error) {
	re, _ := regexp.Compile(`<(.+?)@(.+?)>`) // go home regex, you're drunk!
	if matched := re.FindStringSubmatch(str); len(matched) > 2 {
		host = validHost(matched[2])
		name = matched[1]
	} else {
		if res := strings.Split(name, "@"); len(res) > 1 {
			name = matched[0]
			host = validHost(matched[1])
		}
	}
	if host == "" || name == "" {
		err = errors.New("Invalid address, [" + name + "@" + host + "] address:" + str)
	}
	return name, host, err
}

// Decode strings in Mime header format
// eg. =?ISO-2022-JP?B?GyRCIVo9dztSOWJAOCVBJWMbKEI=?=
func mimeHeaderDecode(str string) string {
	reg, _ := regexp.Compile(`=\?(.+?)\?([QBqp])\?(.+?)\?=`)
	matched := reg.FindAllStringSubmatch(str, -1)
	var charset, encoding, payload string
	if matched != nil {
		for i := 0; i < len(matched); i++ {
			if len(matched[i]) > 2 {
				charset = matched[i][1]
				encoding = strings.ToUpper(matched[i][2])
				payload = matched[i][3]
				switch encoding {
				case "B":
					str = strings.Replace(str, matched[i][0], mailTransportDecode(payload, "base64", charset), 1)
				case "Q":
					str = strings.Replace(str, matched[i][0], mailTransportDecode(payload, "quoted-printable", charset), 1)
				}
			}
		}
	}
	return str
}

func validHost(host string) string {
	host = strings.Trim(host, " ")
	re, _ := regexp.Compile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	if re.MatchString(host) {
		return host
	}
	return ""
}

// decode from 7bit to 8bit UTF-8
// encoding_type can be "base64" or "quoted-printable"
func mailTransportDecode(str string, encoding_type string, charset string) string {
	if charset == "" {
		charset = "UTF-8"
	} else {
		charset = strings.ToUpper(charset)
	}
	if encoding_type == "base64" {
		str = fromBase64(str)
	} else if encoding_type == "quoted-printable" {
		str = fromQuotedP(str)
	}
	if charset != "UTF-8" {
		charset = fixCharset(charset)
		// eg. charset can be "ISO-2022-JP"
		convstr, err := iconv.Conv(str, "UTF-8", charset)
		if err == nil {
			return convstr
		}
	}
	return str
}

func fromBase64(data string) string {
	buf := bytes.NewBufferString(data)
	decoder := base64.NewDecoder(base64.StdEncoding, buf)
	res, _ := ioutil.ReadAll(decoder)
	return string(res)
}

func fromQuotedP(data string) string {
	buf := bytes.NewBufferString(data)
	decoder := qprintable.NewDecoder(qprintable.BinaryEncoding, buf)
	res, _ := ioutil.ReadAll(decoder)
	return string(res)
}

func compress(s string) string {
	var b bytes.Buffer
	w, _ := zlib.NewWriterLevel(&b, zlib.BestSpeed) // flate.BestCompression
	w.Write([]byte(s))
	w.Close()
	return b.String()
}

func fixCharset(charset string) string {
	reg, _ := regexp.Compile(`[_:.\/\\]`)
	fixed_charset := reg.ReplaceAllString(charset, "-")
	// Fix charset
	// borrowed from http://squirrelmail.svn.sourceforge.net/viewvc/squirrelmail/trunk/squirrelmail/include/languages.php?revision=13765&view=markup
	// OE ks_c_5601_1987 > cp949
	fixed_charset = strings.Replace(fixed_charset, "ks-c-5601-1987", "cp949", -1)
	// Moz x-euc-tw > euc-tw
	fixed_charset = strings.Replace(fixed_charset, "x-euc", "euc", -1)
	// Moz x-windows-949 > cp949
	fixed_charset = strings.Replace(fixed_charset, "x-windows_", "cp", -1)
	// windows-125x and cp125x charsets
	fixed_charset = strings.Replace(fixed_charset, "windows-", "cp", -1)
	// ibm > cp
	fixed_charset = strings.Replace(fixed_charset, "ibm", "cp", -1)
	// iso-8859-8-i -> iso-8859-8
	fixed_charset = strings.Replace(fixed_charset, "iso-8859-8-i", "iso-8859-8", -1)
	if charset != fixed_charset {
		return fixed_charset
	}
	return charset
}

func (c *redisClient) redisConnection() (err error) {
	if c.count > 100 {
		c.conn.Close()
		c.count = 0
	}
	if c.count == 0 {
		c.conn, err = redis.Dial("tcp", ":6379")
		if err != nil {
			// handle error
			return err
		}
	}
	return nil
}

func md5hex(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	sum := h.Sum([]byte{})
	return hex.EncodeToString(sum)
}
