import time

import numpy as np
import cv2

try:
    from megapi import *
except ImportError:
    M1=M2=0
    class MegaPi:
        def start(self, x=' '):
            return #nix
        def motorRun(self,x,y):
            return #nix

SLOW = 20
FAST = 70

class MyView:
    WIDTH=320
    HEIGHT=240

    SHOTS = 5
    FRAMES = SHOTS + SHOTS - 1

    SLEEP = 0.2
    
    ca = 180.0/np.pi/2.0

    #camera = None
    #hsv = None
    
    def __init__(self):
        print("--=== 0 ===--")
        self.camera = cv2.VideoCapture(0)
        #self.camera.set(cv2.CAP_PROP_AUTOFOCUS, False)
        self.camera.set(cv2.CAP_PROP_FRAME_WIDTH,self.WIDTH)
        self.camera.set(cv2.CAP_PROP_FRAME_HEIGHT,self.HEIGHT)
        self.cameraOpen(0)
        ret, img = self.camera.read()
        print("Return Code: {}".format(ret))
        self.cameraRelease()
        self.frame = [ np.zeros_like(img) for i in range(0, self.FRAMES) ]
        print("--=== 0 ===--")

        self.hsv = np.zeros_like(img)
        self.hsv[...,1] = 255

        
    def cameraOpen(self, p):
        #self.camera.open(p)
        #self.camera.set(cv2.CAP_PROP_AUTOFOCUS, False)
        #self.camera.set(cv2.CAP_PROP_FRAME_WIDTH,self.WIDTH)
        #self.camera.set(cv2.CAP_PROP_FRAME_HEIGHT,self.HEIGHT)
        x=1
        
    def cameraRelease(self):
        #self.camera.release()
        x=1

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
        font = cv2.FONT_HERSHEY_PLAIN
        fsize = 1
        fcolor = (255,255,255)
        fthick = 1

        lcolor = (255,0,255)
        
        w2 = int(self.WIDTH/2)
        h2 = int(self.HEIGHT/2)
        w4 = int(w2/2)
        h4 = int(h2/2)
        
        y=h4
        for row in res:
            x=w4
            for cell in row:
                cv2.line(bgr,(x,y),(int(x+cell[0]*10),int(y+cell[1]*10)),lcolor,5)
                x+=w2
            y+=h2
        
        cv2.line(bgr,(w2,h2),(int(w2+mx*10),int(h2+my*10)),lcolor,5)
        t = "{}".format(mx)
        cv2.putText(bgr,t,(10,25), font, fsize, fcolor, fthick,cv2.LINE_AA)
        t = "{}".format(my)
        cv2.putText(bgr,t,(10,50), font, fsize, fcolor, fthick,cv2.LINE_AA)
            
        cv2.line(bgr,(int((w2+w4)/2),h2),(int((w2+w4)/2),int(h2+ms*10)),lcolor,5)
        t = "{}".format(ms)
        cv2.putText(bgr,t,(10,75), font, fsize, fcolor, fthick,cv2.LINE_AA)
    

    def oneShot(self,i):
        def decode_fourcc(v):
            v = int(v)
            return "".join([chr((v >> 8 * i) & 0xFF) for i in range(4)])

        while True:
            ret, self.frame[i] = self.camera.read()
            if ret:
                print("Return Code: {}".format(ret))
                break
            print("Retry Shot ...")
            
        fourcc = decode_fourcc(self.camera.get(cv2.CAP_PROP_FOURCC))
        fps = self.camera.get(cv2.CAP_PROP_FPS)
        print(("fourcc {0} fps {1}".format(fourcc,fps)))

            
    def takePictures(self, t=SLEEP):
        self.cameraOpen(0)
        
        for i in range(0, self.FRAMES, 2):
            time.sleep(t)
            self.oneShot(i)
            
        self.cameraRelease()
        

    def encodePictures(self):
        ef = [b' ' for i in range(0,self.FRAMES)]
        for i in range(0, self.FRAMES):
            ret, ef[i] = cv2.imencode('.png', self.frame[i])
        return ef
    

    def writePictures(self):
        for i in range(0, self.FRAMES):
            ret = cv2.imwrite('frame_{0}.png'.format(i), self.frame[i])

            
    def measurePictures(self):
        next = cv2.cvtColor(self.frame[0],cv2.COLOR_BGR2GRAY)

        for i in range(2,self.FRAMES,2):
            prvs = next
            next = cv2.cvtColor(self.frame[i],cv2.COLOR_BGR2GRAY)
            flow = cv2.calcOpticalFlowFarneback(prvs,next, None, 0.5, 3, 15, 3, 5, 1.2, 0)
            self.frame[i-1] = self.flow2hsv(flow)
            (res, mx, my, ms) = self.flow2measure(flow)
            self.showmeasure(self.frame[i-1], res, mx, my, ms)


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
    view.writePictures()
    
    return ret



if __name__ == '__main__':
    bot = MegaPi()
    bot.start()  #'/dev/cu.Makeblock-ELETSPP')
    bot.motorRun(M1,0);
    bot.motorRun(M2,0);

    view = MyView()
    
    while True:
        print("G H J")
        print("V B N")
        
        k = ord(input("Press Key: "))
        if k == 27:
            break
        elif k == ord('g'):
            print(me_control("/me?gapi=forward+left"))
        elif k == ord('h'):
            print(me_control("/me?gapi=forward"))
        elif k == ord('j'):
            print(me_control("/me?gapi=forward+right"))
        elif k == ord('v'):
            print(me_control("/me?gapi=backward+left"))
        elif k == ord('b'):
            print(me_control("/me?gapi=backward"))
        elif k == ord('n'):
            print(me_control("/me?gapi=backward+right"))
        elif k == ord('s'):
            print(me_control("/me?gapi=stop"))
            break

    bot.motorRun(M1,0);
    bot.motorRun(M2,0);

    
