import time
import BaseHTTPServer
import base64

import numpy as np
import cv2

from megapi import *

#M1=M2=0

#class MegaPi:
#    def start(self, x=' '):
#        return #nix
#    def motorRun(self,x,y):
#        return #nix


HOST_NAME = '0.0.0.0'   # 'localhost' # !!!REMEMBER TO CHANGE THIS!!!
PORT_NUMBER = 8088 # Maybe set this to 9000.

SLOW = 20
FAST = 70

class MyView:
    WIDTH=160 #640
    HEIGHT=120 #480
    
    ca = 180.0/np.pi/2.0
    font = cv2.FONT_HERSHEY_SIMPLEX

    #cap = None
    #frame1 = None
    #flow12 = None
    #frame2 = None
    #flow23 = None
    #frame3 = None
    #hsv = None
    
    def __init__(self):
        print("--=== 0 ===--")
        self.cap = cv2.VideoCapture(0)
        self.cap.set(3,self.WIDTH) #width: 640
        self.cap.set(4,self.HEIGHT) #height: 480
        
        ret, frame = self.cap.read()
        self.f1 = np.zeros_like(frame)
        self.f2 = np.zeros_like(frame)
        self.f3 = np.zeros_like(frame)
        ret, self.frame3 = cv2.imencode('.png', frame)
        self.frame1 = np.zeros_like(self.frame3)
        self.flow12 = np.zeros_like(self.frame3)
        self.frame2 = np.zeros_like(self.frame3)
        self.flow23 = np.zeros_like(self.frame3)
        self.hsv = np.zeros_like(frame)
        self.hsv[...,1] = 255

    def flow2hsv(self, flow):
        mag, ang = cv2.cartToPolar(flow[...,0], flow[...,1])
        self.hsv[...,0] = ang*self.ca
        self.hsv[...,2] = cv2.normalize(mag,None,0,255,cv2.NORM_MINMAX)
        bgr = cv2.cvtColor(self.hsv,cv2.COLOR_HSV2BGR)

        return bgr


    def flow2measure(self, flow):
        height, width = flow.shape[:2]
        res = cv2.resize(flow, None, fx=2.0/width, fy=2.0/height, interpolation = cv2.INTER_AREA)
        
        mx = (res[0,0,0]+res[0,1,0]+res[1,0,0]+res[1,1,0])/4.0
        my = (res[0,0,1]+res[0,1,1]+res[1,0,1]+res[1,1,1])/4.0
    
        ms = (res[0,0,0]-res[0,1,0]+res[1,0,0]-res[1,1,0]+
              res[0,0,1]+res[0,1,1]-res[1,0,1]-res[1,1,1])/8.0

        return res, mx, my, ms

    
    def showmeasure(self, bgr, res, mx, my, ms):
        w2 = self.WIDTH/2
        h2 = self.HEIGHT/2
        w4 = w2/2
        h4 = h2/2
        
        y=h4
        for row in res:
            x=w4
            for cell in row:
                cv2.line(bgr,(x,y),(int(x+cell[0]*10),int(y+cell[1]*10)),(255,0,255),5)
                x+=w2
            y+=h2
        
        cv2.line(bgr,(w2,h2),(int(w2+mx*10),int(h2+my*10)),(255,0,255),5)
        t = "{}".format(abs(mx))
        cv2.putText(bgr,t,(10,25), self.font, 1,(255,255,255),2,cv2.LINE_AA)
        t = "{}".format(abs(my))
        cv2.putText(bgr,t,(10,50), self.font, 1,(255,255,255),2,cv2.LINE_AA)
            
        cv2.line(bgr,((w2+w4)/2,h2),((w2+w4)/2,int(h2+ms*10)),(255,0,255),5)
        t = "{}".format(abs(ms))
        cv2.putText(bgr,t,(10,75), self.font, 1,(255,255,255),2,cv2.LINE_AA)
    
    
    def takePictures(self):
        ret, self.f1 = self.cap.read()
        time.sleep(0.5)
        ret, self.f2 = self.cap.read()
        time.sleep(0.5)
        ret, self.f3 = self.cap.read()
    
    def measurePictures(self):
        prvs = cv2.cvtColor(self.f1,cv2.COLOR_BGR2GRAY)
        ret, self.frame1 = cv2.imencode('.png', self.f1)
        ret = cv2.imwrite('frame1.png', self.f1)

        next = cv2.cvtColor(self.f2,cv2.COLOR_BGR2GRAY)
        flow = cv2.calcOpticalFlowFarneback(prvs,next, None, 0.5, 3, 15, 3, 5, 1.2, 0)
        bgr = self.flow2hsv(flow)
        (res, mx, my, ms) = self.flow2measure(flow)
        self.showmeasure(bgr, res, mx, my, ms)
        ret, self.flow12 = cv2.imencode('.png', bgr)
        ret = cv2.imwrite('flow12.png', self.bgr)
        ret, self.frame2 = cv2.imencode('.png', self.f2)
        ret = cv2.imwrite('frame2.png', self.f2)

        prvs = next

        next = cv2.cvtColor(self.f3,cv2.COLOR_BGR2GRAY)
        flow = cv2.calcOpticalFlowFarneback(prvs,next, None, 0.5, 3, 15, 3, 5, 1.2, 0)
        bgr = self.flow2hsv(flow)
        (res, mx, my, ms) = self.flow2measure(flow)
        self.showmeasure(bgr, res, mx, my, ms)
        ret, self.flow23 = cv2.imencode('.png', bgr)
        ret = cv2.imwrite('flow23.png', self.bgr)
        ret, self.frame3 = cv2.imencode('.png', self.f3)
        ret = cv2.imwrite('frame3.png', self.f3)


def me_control(path):
    (pre,sep,post) = path.partition("?")
    (pre,sep,post) = post.partition("=")

    if pre != "gapi":
        ret = "Error: Missing gapi, found {0}".format(pre)
    else:
        ret = "Error: Don't understand {0}".format(post)

        if post == "forward+left":
            bot.motorRun(M1,-SLOW)
            bot.motorRun(M2,FAST)
            ret = "Go forward and left!"

        if post == "forward":
            bot.motorRun(M1,-FAST);
            bot.motorRun(M2,FAST);
            ret = "Go forward!"
    
        if post == "forward+right":
            bot.motorRun(M1,-FAST);
            bot.motorRun(M2,SLOW);
            ret = "Go forward and right!"
    
        if post == "backward+left":
            bot.motorRun(M1,SLOW);
            bot.motorRun(M2,-FAST);
            ret = "Go backward and left!"
    
        if post == "backward":
            bot.motorRun(M1,FAST);
            bot.motorRun(M2,-FAST);
            ret = "Go backward!"
    
        if post == "backward+right":
            bot.motorRun(M1,FAST);
            bot.motorRun(M2,-SLOW);
            ret = "Go backward and right!"
    
        if post == "stop":
            bot.motorRun(M1,0);
            bot.motorRun(M2,0);
            ret = "Stop!"    

    view.takePictures()
    bot.motorRun(M1,0);
    bot.motorRun(M2,0);
    view.measurePictures()

    return ret


def b64img(png):
    spng = png.tostring()
    bpng = bytes(base64.b64encode(spng))
    return b"<p><img src='data:image/png;base64," + bpng + b"'/></p>"

        
def cv_pictures():
    html = b64img(view.frame1) + b64img(view.flow12) + b64img(view.frame2) + b64img(view.flow23) + b64img(view.frame3)
    return html


class MyHandler(BaseHTTPServer.BaseHTTPRequestHandler):
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
        s.wfile.write(s.CONTROL)
        s.wfile.write("<p>MEGAPI: %s</p>" % me_control(s.path))
        s.wfile.write(cv_pictures())
        s.wfile.write("</body></html>")

if __name__ == '__main__':
    bot = MegaPi()
    bot.start()  #'/dev/cu.Makeblock-ELETSPP')
    bot.motorRun(M1,0);
    bot.motorRun(M2,0);

    view = MyView()
    
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
    
    view.cap.release()
    
