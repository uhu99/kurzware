import time
import BaseHTTPServer

from megapi import *

HOST_NAME = 'localhost' # !!!REMEMBER TO CHANGE THIS!!!
PORT_NUMBER = 8088 # Maybe set this to 9000.

CONTROL = """
    <p>
      <form method="get" id="me" action="/me">
        <table>
          <tr>
            <td align="left"><input type="submit" name="gapi" value="forward left" /></td>
            <td align="center"><input type="submit" name="gapi" value="forward" /></td>
            <td align="right"><input type="submit" name="gapi" value="forward right" /></td>
          </tr>
          <tr>
            <td align="left"><input type="submit" name="gapi" value="stop" /></td>
            <td align="center"><input type="submit" name="gapi" value="stop" /></td>
            <td align="right"><input type="submit" name="gapi" value="stop" /></td>
          </tr>
          <tr>
            <td align="left"><input type="submit" name="gapi" value="backward left" /></td>
            <td align="center"><input type="submit" name="gapi" value="backward" /></td>
            <td align="right"><input type="submit" name="gapi" value="backward right" /></td>
          </tr>
        </table>
      </form>
    </p>
"""

SLOW = 20
FAST = 70

def me_control(path):
    (pre,sep,post) = path.partition("?")
    (pre,sep,post) = post.partition("=")

    if pre != "gapi":
        return "Error: Missing gapi, found {0}".format(pre)

    if post == "forward+left":
        bot.motorRun(M1,-SLOW)
        bot.motorRun(M2,FAST)
        return "Go forward and left!"

    if post == "forward":
        bot.motorRun(M1,-FAST);
        bot.motorRun(M2,FAST);
        return "Go forward!"
    
    if post == "forward+right":
        bot.motorRun(M1,-FAST);
        bot.motorRun(M2,SLOW);
        return "Go forward and right!"
    
    if post == "backward+left":
        bot.motorRun(M1,SLOW);
        bot.motorRun(M2,-FAST);
        return "Go backward and left!"
    
    if post == "backward":
        bot.motorRun(M1,FAST);
        bot.motorRun(M2,-FAST);
        return "Go backward!"
    
    if post == "backward+right":
        bot.motorRun(M1,FAST);
        bot.motorRun(M2,-SLOW);
        return "Go backward and right!"
    
    if post == "stop":
        bot.motorRun(M1,0);
        bot.motorRun(M2,0);
        return "Stop!"    
        
    return "Error: Don't understand {0}".format(post)


class MyHandler(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_HEAD(s):
        s.send_response(200)
        s.send_header("Content-type", "text/html")
        s.end_headers()
    def do_GET(s):
        """Respond to a GET request."""
        s.send_response(200)
        s.send_header("Content-type", "text/html")
        s.end_headers()
        s.wfile.write("<html><head><title>Title goes here.</title></head>")
        s.wfile.write("<body><p>This is a test.</p>")
        # If someone went to "http://something.somewhere.net/foo/bar/",
        # then s.path equals "/foo/bar/".
        s.wfile.write("<p>You accessed path: %s</p>" % s.path)
        s.wfile.write(CONTROL)
        s.wfile.write("<p>MEGAPI: %s</p>" % me_control(s.path))
        s.wfile.write("</body></html>")

if __name__ == '__main__':
    bot = MegaPi()
    bot.start('/dev/cu.Makeblock-ELETSPP')
    bot.motorRun(M1,0);
    bot.motorRun(M2,0);
    server_class = BaseHTTPServer.HTTPServer
    httpd = server_class((HOST_NAME, PORT_NUMBER), MyHandler)
    print time.asctime(), "Server Starts - %s:%s" % (HOST_NAME, PORT_NUMBER)
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass
    httpd.server_close()
    print time.asctime(), "Server Stops - %s:%s" % (HOST_NAME, PORT_NUMBER)
    bot.motorRun(M1,0);
    bot.motorRun(M2,0);
    
