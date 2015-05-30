/* -*- c-mode -*- */
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netdb.h>

#define	VOLTAGE_ON  0x41
#define	VOLTAGE_OFF 0x82

#define	SWITCH_ON   0x01
#define	SWITCH_OFF  0x02
#define	SWITCH_NONE 0x04

#define	SCHEDULE_OFF_ONCE   0x00
#define	SCHEDULE_ON_ONCE    0x01
#define	SCHEDULE_OFF_PERIOD 0x02
#define	SCHEDULE_ON_PERIOD  0x03
#define	SCHEDULE_LOOP       0xe5

#define SCHEDULE_SOCKET "\x80\x01\x02\x03\x04"

typedef unsigned char uint8;

int DEBUG = 1; // true/false

char HOST[] = "192.168.178.6";
char PORT[] = "5000";
int sock;

char key[]  = "1       ";

uint8 task[4] = {0,0,0,0};
uint8 stat[4] = {0,0,0,0};
uint8 statcrypt[4] = {0,0,0,0};

// Usage Information
void usage()
{
  printf("Usage:");
  printf("--host ip       Host IP address");
  printf("--port n        IP port");
  printf("--pw xxx        Password");
  printf("--socket n      Socket <n> (1..4)");
  printf("--status        Status of sockets");
  printf("--on            Switch on socket <n>");
  printf("--off           Switch off socket <n>");
  printf("--control bbbb  Control all sockets, b=(0,1,x)");
  printf("--read          Read schedule of socket <n>");
  printf("--write json    Write schedule of socket <n>");
  printf("--debug         Print Debug information");

  exit(1);
}


// Print Debug Information
void dprint(const char *x, ...)
{
  va_list ap;
  if (DEBUG) {
    printf(x, ap);
  }
}


//  Get Status of Sockets
uint8* status()
{
  int len = recv(sock, statcrypt, 4, 0);
  if (!len) exit(1);

  stat[3]=(((statcrypt[0]-key[1])^key[0])-task[3])^task[2];
  stat[2]=(((statcrypt[1]-key[1])^key[0])-task[3])^task[2];
  stat[1]=(((statcrypt[2]-key[1])^key[0])-task[3])^task[2];
  stat[0]=(((statcrypt[3]-key[1])^key[0])-task[3])^task[2];

  printf ("\nReceived: %02x%02x%02x%02x %02x%02x%02x%02x\n",
            statcrypt[0], statcrypt[1], statcrypt[2], statcrypt[3], 
            stat[0], stat[1], stat[2], stat[3]);

  return stat;
}


/******************************************************************
  Print verbose Status
  
  Arguments:
  - `stat`: bytearray(4)   raw Status
******************************************************************/
void print_status() 
{
  int i;
  for (i=0;i<4;i++) {
    printf("\n%d: ", i);
    switch(stat[i]) {
    case VOLTAGE_ON: printf("ON"); break;
    case VOLTAGE_OFF: printf("OFF"); break;
    default: printf("???");
    }
  }
}

//  Setup connection to EG-PMS
uint8* myconnect() {
  uint8 x11 = 0x11;
  uint8 res[4] = {0,0,0,0};
  unsigned int res10, res32;
  int err;
  struct addrinfo hints, *raddr;


  // first, load up address structs with getaddrinfo():
  memset(&hints, 0, sizeof hints);
  hints.ai_family = AF_UNSPEC;
  hints.ai_socktype = SOCK_STREAM;
  err = getaddrinfo(HOST, PORT, &hints, &raddr);
  if (err) exit(1);

  // make a socket:
  sock = socket(raddr->ai_family, raddr->ai_socktype, raddr->ai_protocol);
  if (sock < 0) exit(1);

  // connect!
  err = connect(sock, raddr->ai_addr, raddr->ai_addrlen);
  if (err < 0) exit(1);

  // sock.SetDeadline(time.Now().Add(time.Second + time.Second))

  err = send(sock, &x11, 1, 0);
  if (err < 0) exit(1);

  err = recv(sock, &task, 4, 0);
  if (err < 0) exit(1);
  printf ("\nReceived: %x %x %x %x", task[0], task[1], task[2], task[3]);

  res10 = ((task[0]^key[2])*((unsigned int)key[0]))^(key[6]|(((unsigned int)key[4])<<8))^task[2];
  res32 = ((task[1]^key[3])*((unsigned int)key[1]))^(key[7]|(((unsigned int)key[5])<<8))^task[3];

  res[0] = res10 & 0xff;
  res[1] = (res10 >> 8) & 0xff;
  res[2] = res32 & 0xff;
  res[3] = (res32 >> 8) & 0xff;

  err = send(sock, &res, 4, 0);
  if (err < 0) exit(1);
  printf ("\nSolution: %02x%02x%02x%02x", res[0], res[1], res[2], res[3]);

  return status();
}


/******************************************************************
  Switch Sockets on/off
  Arguments:
  - `bbbb`: SWITCH_ON/OFF/NONE of Sockets 1/2/3/4
******************************************************************/
uint8* control(char* bbbb) {
  uint8 ctrl[4] = {0,0,0,0};
  uint8 ctrlcrypt[4] = {0,0,0,0};
int i, err;

  for (i=0; i<4; i++)
    switch (bbbb[i]) {
    case SWITCH_ON:
    case SWITCH_OFF:
    case SWITCH_NONE: ctrl[i] = bbbb[i]; break;
    case '0': ctrl[i] = SWITCH_OFF; break;
    case '1': ctrl[i] = SWITCH_ON; break;
 default: ctrl[i] = SWITCH_NONE;
}

  ctrlcrypt[0]=((((ctrl[3]^task[2])+task[3])^key[0])+key[1]) & 0xff;
  ctrlcrypt[1]=((((ctrl[2]^task[2])+task[3])^key[0])+key[1]) & 0xff;
  ctrlcrypt[2]=((((ctrl[1]^task[2])+task[3])^key[0])+key[1]) & 0xff;
  ctrlcrypt[3]=((((ctrl[0]^task[2])+task[3])^key[0])+key[1]) & 0xff;

  err = send(sock, ctrlcrypt, 4, 0);
  if (err < 0) exit(1);
  printf ("\nControl: %02x%02x%02x%02x %02x%02x%02x%02x",
          ctrl[0], ctrl[1], ctrl[2], ctrl[3],
          ctrlcrypt[0], ctrlcrypt[1], ctrlcrypt[2], ctrlcrypt[3]);

  return status();
}


//  Close Connection
void myclose() {
  // Sende ein EOF-Byte, sodass der Server beim nächsten read() als Rückgabewert 0 bekommt
  //und die Verbindung beenden kann
  shutdown(sock,SHUT_WR);
 
  close(sock);
}


//  EnerGenie Client
int main(int argc, char *argv[]) {
  if ((argc == 2) && (strlen(argv[1]) == 4)) {
    printf("%d %s", argc, argv[1]);
    myconnect();
    print_status();
    control(argv[1]);
    print_status();
    myclose();
  } else usage();

  return 0;
}

