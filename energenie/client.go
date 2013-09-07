package main

import (
	"fmt"
	"os"
	"net"
	"time"
	"strconv"
	"encoding/json"
)

const (
	VOLTAGE_ON  = 0x41
	VOLTAGE_OFF = 0x82
)

const (
	SWITCH_ON   = 0x01
	SWITCH_OFF  = 0x02
	SWITCH_NONE = 0x04
)

const (
	SCHEDULE_OFF_ONCE   = 0x00
	SCHEDULE_ON_ONCE    = 0x01
	SCHEDULE_OFF_PERIOD = 0x02
	SCHEDULE_ON_PERIOD  = 0x03
	SCHEDULE_LOOP       = 0xe5
)

const SCHEDULE_SOCKET = "\x80\x01\x02\x03\x04"

//"%Y-%m-%dT%H:%M:%S"
const DATETIME_FORMAT = "2006-01-02T15:04:00" 


type Exec struct {
	Switch string
	Time time.Time
}


type Schedule struct {
	Setup time.Time
	Lines []Exec
	Loop string  //time.Duration
}


var DEBUG = false

var HOST, PORT = "192.168.178.6", 5000
//var HOST, PORT = "192.168.0.254", 5000

var key  = []byte("1       ")

var task = []byte{0,0,0,0}

var raddr *net.TCPAddr
var sock *net.TCPConn

// Usage Information
func usage() {
	fmt.Println("Usage:")
	fmt.Println("--host ip       Host IP address")
	fmt.Println("--port n        IP port")
	fmt.Println("--pw xxx        Password")
	fmt.Println("--socket n      Socket <n> (1..4)")
	fmt.Println("--status        Status of sockets")
	fmt.Println("--on            Switch on socket <n>")
	fmt.Println("--off           Switch off socket <n>")
	fmt.Println("--control bbbb  Control all sockets, b=(0,1,x)")
	fmt.Println("--read          Read schedule of socket <n>")
	fmt.Println("--write json    Write schedule of socket <n>")
	fmt.Println("--debug         Print Debug information")

	os.Exit(1)
}


// Print Debug Information
func dprint(x ...interface{}) {
	if DEBUG {
		fmt.Println(x)
	}
}


//  Get Status of Sockets
func status() []byte {
	var stat = []byte{0,0,0,0}
	var statcrypt = []byte{0,0,0,0}

	_,err := sock.Read(statcrypt)
	perror(err)
	stat[3]=byte(((((uint(statcrypt[0])-uint(key[1]))^uint(key[0]))-uint(task[3]))^uint(task[2])) & 0xff)
	stat[2]=byte(((((uint(statcrypt[1])-uint(key[1]))^uint(key[0]))-uint(task[3]))^uint(task[2])) & 0xff)
	stat[1]=byte(((((uint(statcrypt[2])-uint(key[1]))^uint(key[0]))-uint(task[3]))^uint(task[2])) & 0xff)
	stat[0]=byte(((((uint(statcrypt[3])-uint(key[1]))^uint(key[0]))-uint(task[3]))^uint(task[2])) & 0xff)
	dprint ("Received: ", statcrypt, stat)

	return stat
}


/******************************************************************
  Print verbose Status
  
  Arguments:
  - `stat`: bytearray(4)   raw Status
******************************************************************/
func print_status(stat []byte) {
	s := ""
	for i,b := range stat {
		if b == VOLTAGE_ON {
			s += "1"
			dprint (i, ": ON")
		} else if b == VOLTAGE_OFF {
			s += "0"
			dprint (i, ": OFF")
		} else {
			s += "?"
			dprint (i, ": ???")
		}
	}

	fmt.Println(s)
}


func perror(err interface{error}) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

//  Setup connection to EG-PMS
func connect() []byte {
	var res = make([]byte, 4)
	var res10, res32 uint
	var err error

	host := HOST + ":" + strconv.FormatInt(int64(PORT), 10)
	dprint("Host: ", host)
	raddr,_ = net.ResolveTCPAddr("tcp4", host)
	sock,err = net.DialTCP("tcp4", nil, raddr)
	perror(err)
	sock.SetDeadline(time.Now().Add(time.Second + time.Second))

	_,err = sock.Write([]byte{0x11})
	perror(err)

	_,err = sock.Read(task)
	perror(err)
	dprint ("Received: ", task)

	res10 = ((uint(task[0])^uint(key[2]))*uint(key[0]))^(uint(key[6])|(uint(key[4])<<8))^uint(task[2])
	res32 = ((uint(task[1])^uint(key[3]))*uint(key[1]))^(uint(key[7])|(uint(key[5])<<8))^uint(task[3])

	res[0] = byte(res10 & 0xff)
	res[1] = byte((res10 >> 8) & 0xff)
	res[2] = byte(res32 & 0xff)
	res[3] = byte((res32 >> 8) & 0xff)

	_,err = sock.Write(res)
	perror(err)
	dprint ("Solution: ", res)

	return status()
}


/******************************************************************
  Switch Sockets on/off
  Arguments:
  - `bbbb`: SWITCH_ON/OFF/NONE of Sockets 1/2/3/4
******************************************************************/
func control(bbbb []byte) []byte {
	var ctrl = make([]byte, 4)
	var ctrlcrypt = make([]byte, 4)

	for i,v := range bbbb {
		switch v {
		case SWITCH_ON: fallthrough
		case SWITCH_OFF: fallthrough
		case SWITCH_NONE: ctrl[i] = v
		case '0': ctrl[i] = SWITCH_OFF
		case '1': ctrl[i] = SWITCH_ON
		default: ctrl[i] = SWITCH_NONE
		}
	}

	ctrlcrypt[0]=((((ctrl[3]^task[2])+task[3])^key[0])+key[1]) & 0xff
	ctrlcrypt[1]=((((ctrl[2]^task[2])+task[3])^key[0])+key[1]) & 0xff
	ctrlcrypt[2]=((((ctrl[1]^task[2])+task[3])^key[0])+key[1]) & 0xff
	ctrlcrypt[3]=((((ctrl[0]^task[2])+task[3])^key[0])+key[1]) & 0xff
	_,err := sock.Write(ctrlcrypt)
	perror(err)
	dprint ("Control:  ", ctrl, ctrlcrypt)

	return status()
}


/******************************************************************
  Read Schedule of Socket i
  
  Arguments:
  - `i`: Number of Socket
******************************************************************/
func read_schedule(i int) []byte {
	sch := []byte{0x00,0x00,0x00,0x00}
	sch = append(sch, 0xe5)
	sch = append(sch, 0x00,0x00,0x00,0x00)
	sch = append(sch, 0x00)
	sch = append(sch, 0x00,0x00)

	if (i < 1) || (i > 4) {
		return sch
	}

	data := SCHEDULE_SOCKET[i] | SCHEDULE_SOCKET[0]
	dprint ("Socket: ", data)
	sch[9] = data

	return send_schedule(sch)
}


func send_schedule(sch []byte) []byte {
	var schcrypt []byte
	var i, j int
	var length, checksum int
	var err error

	length = len(sch)
	checksum = 0

	for _,b := range sch {
		checksum -= int(b)
	}
	sch[length-2] = byte(checksum & 0xff)
	sch[length-1] = byte((checksum >> 8) & 0xff)

	schcrypt = make([]byte, length)
	j = 0
	for i=length; i!=0; {
		i -= 1
		schcrypt[i] = ((((sch[j]^task[2])+task[3])^key[0])+key[1]) & 0xff
		j += 1
	}
	dprint ("Schedule:  ", length, sch, schcrypt)
	_,err = sock.Write(schcrypt)
	perror(err)

	// Decrypt answer

	schcrypt = make([]byte, 250)
	length,err = sock.Read(schcrypt)
	perror(err)
	sch = make([]byte, length)
	j = 0
	for i=length; i!=0; {
		i -= 1
		sch[i] = ((((schcrypt[j]-key[1])^key[0])-task[3])^task[2]) & 0xff
		j += 1
	}
	dprint ("Schedule:  ", length, sch, schcrypt)

	return sch
}

//  Close Connection
func close() {
	sock.Close()
}


/******************************************************************
  Write verbose schedule
  Arguments:
  - `sch`: raw Schedule
******************************************************************/
func print_schedule(sch []byte) {
	var jsched Schedule
	var exec []Exec
	var i, j int
	length := len(sch)

	tsetup := bytes2int(sch[0:4])
	dsetup := time.Unix(tsetup,0)
	dprint ("Setup: ", dsetup)
	jsched.Setup = dsetup //.Format(DATETIME_FORMAT)}

	max := length - 12
	exec = make([]Exec, max/5, 50)
	j = 0
	for i=4; sch[i] != SCHEDULE_LOOP; i+=5 {
		switch sch[i] {
		case SCHEDULE_OFF_ONCE: exec[j].Switch = "OFF_ONCE"
		case SCHEDULE_ON_ONCE:  exec[j].Switch = "ON_ONCE"
		case SCHEDULE_OFF_PERIOD: exec[j].Switch = "OFF_PERIOD"
		case SCHEDULE_ON_PERIOD:  exec[j].Switch = "ON_PERIOD"
		default: exec[j].Switch = "???"
		}

		texec := bytes2int(sch[i+1:i+5])
		dexec := time.Unix(texec,0)
		dprint ("Entry: ", j, sch[i], dexec)
		exec[j].Time = dexec //.Format(DATETIME_FORMAT)
		j+=1
	}
	jsched.Lines = exec[0:j]

	tperiod := bytes2int(sch[i+1:i+5])
	dperiod := time.Duration(tperiod) * time.Second
	dprint ("Loop: ", sch[i], dperiod, sch[i+6])
	jsched.Loop = dperiod.String()

	b,_ := json.Marshal(jsched)
	fmt.Println(string(b))
}

/******************************************************************
  Convert bytearray (4 bytes) to integer
  
  Arguments:
  - `b`: bytearray
******************************************************************/
func bytes2int(b []byte) int64 {
	i := int64(b[3])
	i = (i << 8) | int64(b[2])
	i = (i << 8) | int64(b[1])
	i = (i << 8) | int64(b[0])

	return i
}


/******************************************************************
  Convert integer (32 bit) in bytearray
  
  Arguments:
  - `i`: integer
******************************************************************/
func int2bytes(i int64) []byte {
	var b = make([]byte, 4)

	b[0] = byte(i & 0xff)
	i >>= 8
	b[1] = byte(i & 0xff)
	i >>= 8
	b[2] = byte(i & 0xff)
	i >>= 8
	b[3] = byte(i & 0xff)

	return b
}


/******************************************************************
  Write schedule for Socket s
  
  Arguments:
  - `s`: Socket number
  - `w`: Schedule string
******************************************************************/
func write_schedule(s int, w string) []byte {
	var sch []byte
	var jschedule Schedule
	var tperiod time.Duration
	var err error

	err = json.Unmarshal([]byte(w), &jschedule)
	perror(err)

	fmt.Println(jschedule)

	sch = int2bytes(jschedule.Setup.Unix())

	for i,e := range jschedule.Lines {
		switch e.Switch {
		case "OFF_ONCE": sch = append(sch, SCHEDULE_OFF_ONCE)
		case "ON_ONCE":  sch = append(sch, SCHEDULE_ON_ONCE)
		case "OFF_PERIOD": sch = append(sch, SCHEDULE_OFF_PERIOD)
		case "ON_PERIOD":  sch = append(sch, SCHEDULE_ON_PERIOD)
		default:
			fmt.Printf("Unknown Schedule attribute %d: '%s'\n", i, e.Switch)
			os.Exit(1)
		}
		sch = append(sch, 0, 0, 0, 0)
		copy(sch[len(sch)-4:], int2bytes(e.Time.Unix()))
	}

	sch = append(sch, SCHEDULE_LOOP)
	dprint ("Loop Period: ", jschedule.Loop)
	sch = append(sch, 0, 0, 0, 0)
	tperiod,err = time.ParseDuration(jschedule.Loop)
	perror(err)
	copy(sch[len(sch)-4:], int2bytes(int64(tperiod.Seconds())))

	sch = append(sch, SCHEDULE_SOCKET[s])

	// Checksum placeholder
	sch = append(sch, 0x00, 0x00)

	if DEBUG { print_schedule(sch) }

	return send_schedule(sch)
}

func test() {
	var s = 2
	//var w = "{\"Setup\":\"2013-09-04T13:18:00+02:00\",\"Lines\":[{\"Switch\":\"ON_PERIOD\",\"Time\":\"2013-09-04T15:22:21+02:00\"},{\"Switch\":\"OFF_PERIOD\",\"Time\":\"2013-09-04T15:23:22+02:00\"},{\"Switch\":\"OFF_ONCE\",\"Time\":\"2013-08-05T14:03:03+02:00\"},{\"Switch\":\"ON_ONCE\",\"Time\":\"2013-08-05T14:04:04+02:00\"}],\"Loop\":\"13h11m7s\"}"
	var w = "{\"Setup\":\"2013-09-04T13:18:00+02:00\",\"Lines\":[{\"Switch\":\"ON_PERIOD\",\"Time\":\"2013-09-04T15:22:21+02:00\"},{\"Switch\":\"OFF_PERIOD\",\"Time\":\"2013-09-04T15:23:22+02:00\"},{\"Switch\":\"OFF_ONCE\",\"Time\":\"2013-08-05T14:03:03+02:00\"},{\"Switch\":\"ON_ONCE\",\"Time\":\"2013-08-05T14:04:04+02:00\"}],\"Loop\":\"2d13h11m7s\"}"

	fmt.Println("Hallo, Welt!")

	print_status(connect())
	print_status(control([]byte("1xxx")))
	print_schedule(write_schedule(s,w))
	close()
}

//  EnerGenie Client
func main() {
	var Usage string
	var Socket int
	var Status bool
	var OnOff byte
	var Read bool
	var Write string
	var bbbb   []byte
	var st []byte
	var err error
	var i64 int64

	Usage  = ""
	Socket = 0
	Status = false
	OnOff  = SWITCH_NONE
	Read   = false
	Write  = ""

	length := len(os.Args)
	if length < 2 {
		fmt.Println(os.Args[0])
		usage()
	}

	for i:=1; i<length; i++ {
		a := os.Args[i]
		switch a {
		case "-h", "--help":
			Usage = "energenie_client V1.0"
		case "--host":
			i++
			if i >= length {
				Usage = "Too few arguments"
				break
			}
			HOST = os.Args[i]
		case "--port":
			i++
			if i >= length {
				Usage = "Too few arguments"
				break
			}
			i64,err = strconv.ParseInt(os.Args[i], 0, 0)
			PORT = int(i64)
			perror(err)
		case "--pw":
			i++
			if i >= length {
				Usage = "Too few arguments"
				break
			}
			copy(key, []byte(os.Args[i]))
		case "--socket":
			i++
			if i >= length {
				Usage = "Too few arguments"
				break
			}
			i64,err = strconv.ParseInt(os.Args[i], 0, 0)
			Socket = int(i64)
			perror(err)
			if Socket < 1 || Socket > 4 {
				Usage = "Wrong socket number " + os.Args[i]
				break
			}
		case "--status":
			Status = true
		case "--on":
			OnOff = SWITCH_ON
		case "--off":
			OnOff = SWITCH_OFF
		case "--control":
			i++
			if i >= length {
				Usage = "Too few arguments"
				break
			}
			bbbb = []byte(os.Args[i])
			if len(bbbb) != 4 {
				Usage = "Wrong number of control bits " + string(bbbb)
				break
			}
		case "--read":
			Read = true
		case "--write":
			i++
			Write = os.Args[i]
		case "--debug":
			DEBUG = true
		default:
			Usage = "Unhandled option " + a
			break
		}
	}

	if len(Usage) > 0 {
		fmt.Println(Usage)
		usage()
	}

	if OnOff != SWITCH_NONE && Socket > 0 {
		st  = connect()
		st[Socket-1] = OnOff
		st = control(st)
	} else if 0 < len(bbbb) {
		if 0 == len(st) {
			st = connect()
		}
		st = control(bbbb)
	}

	if Read && Socket > 0 {
		if 0 == len(st) {
			st = connect()
			st = control([]byte("xxxx"))
		}
		print_schedule(read_schedule(Socket))
	}

	if 0 < len(Write) && !Read && Socket > 0 {
		if 0 == len(st) {
			st = connect()
			st = control([]byte("xxxx"))
		}
		print_schedule(write_schedule(Socket, Write))
	}

	if Status {
		if 0 == len(st) {
			st = connect()
		}
		print_status(st)
	}

	if 0 < len(st) {
		close()
	}
}
