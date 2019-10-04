package main
 
import (
 
	"fmt"
	"os"
        "os/exec"
        "syscall"
	"io"
        "strconv"
        "strings"
        "regexp"
        "bufio"
        "time"
        "flag"
        //"C"
)
 
type searchProcess struct {
  programName string
  processString string
  paramToAdd string
  status int
}
 
type Process struct {
pid   int
programName string
startTime string
elapseTime string
cpuUsage string
memoryUsage string
status int
}
 
var (
config[100] searchProcess
gcheck[100] Process
procCnt,confCnt int = 0,0
WHITE, RED, GREEN string = "\033[0m", "\033[0;31m", "\033[0;32m"
CLOCK_TICK, PAGE_SIZE, MEM_TOTAL uint64
)
 
/*
func GetClockTicksPerSecond() (ticksPerSecond uint64) {
  var sc_clk_tck C.long
  sc_clk_tck = C.sysconf(C._SC_CLK_TCK)
  ticksPerSecond = uint64(sc_clk_tck)
  return
}
*/
 
func getClockTicksPerSecond() {
  out, err := exec.Command("/usr/bin/getconf", "CLK_TCK").Output()
  // ignore errors
  if err == nil {
    i, err := strconv.ParseInt(strings.Trim(string(out),"\n"),10, 64)
    if err == nil {
      CLOCK_TICK = uint64(i)
    }
  }
}
 
func getPageSize() {
  out, err := exec.Command("/usr/bin/getconf", "PAGESIZE").Output()
  // ignore errors
  if err == nil {
    i, err := strconv.ParseInt(strings.Trim(string(out),"\n"),10, 64)
    if err == nil {
      PAGE_SIZE = uint64(i)
    }
  }
  return
}
 
 
 
//get a nice timestamp for the elapse time
func convertTimestamp(timestamp uint64)(string) {
 
  var (
   day,hour,min,sec uint64
   elapseTime uint64
   etime string
  )
 
  curtime := time.Now()
  cursec := curtime.Unix()
  elapseTime = uint64(cursec) - timestamp
 
  day = elapseTime / 86400
  if( day >= 1 ) {
    hour = (elapseTime - (day * 86400)) / 3600
  }else{
    hour = elapseTime / 3600
  }
 
  if( hour >= 1 ) {
    min = (elapseTime - ((day * 86400) + (hour * 3600))) / 60
  }else{
    min =  (elapseTime - (day * 86400)) / 60
  }
 
  if ( min >= 1 ) {
    sec = elapseTime - ((day * 86400) + (hour * 3600) + (min * 60))
  }else{
    sec = elapseTime - (day * 86400) + (hour * 3600)
  }
 
  etime = fmt.Sprintf("%d-%02d:%02d:%02d",day,hour,min,sec)
  return etime 
}
 
//parsing the configuration file
func readConfig(configfile string)(int,error) {
 
  c, err := os.Open(configfile)
  if err != nil {
    return 1,err
  }
  defer c.Close()
 
  scanner := bufio.NewScanner(c)
  for scanner.Scan() {
    cnf := strings.Split(scanner.Text(),";") 
    config[confCnt].programName = cnf[0]
    config[confCnt].processString = cnf[1]
    config[confCnt].paramToAdd = cnf[2]
    config[confCnt].status = 0
  //  fmt.Printf("%s - %s - %s\n",config[confCnt].programName,config[confCnt].processString,config[confCnt].paramToAdd)
    confCnt++
  } 
  return 0,nil
}
 
// print on the console
func printProcess() {
 
  var (
    configLen,gcheckLen,progNameLen,tmp,nbdot,y int
    pid,dotty,status string
  )
 
  gcheckLen = procCnt
  configLen = confCnt
 
  progNameLen=0
  for y=0;y<gcheckLen;y++ {
    tmp = len(gcheck[y].programName)
    if(tmp > progNameLen ) {
      progNameLen = tmp
    } 
  }
  progNameLen=progNameLen+5
 
  // print header
  nbdot=progNameLen-7
  dotty = fmt.Sprintf("%*s",nbdot," ")
  fmt.Printf("Process %s   %-11s%-8s%-8s%-20s%-15sSTATUS\n",dotty,"PID","%CPU","%MEM","S_TIME","E_TIME")
 
  // print body
  for y=0;y<gcheckLen;y++ {
    tmp=len(gcheck[y].programName)
    nbdot=progNameLen-tmp
    dotty = fmt.Sprintf("%*s",nbdot," ")
    dotty = strings.Replace(dotty," ",".",-1)
    pid="-"
    /*if (gcheck[y].status==0 ) { 
      status=fmt.Sprintf("%s[ KO ]%s",RED,WHITE)
    }else{*/
    status=fmt.Sprintf("%s[ OK ]%s",GREEN,WHITE)
    pid=strconv.Itoa(gcheck[y].pid)
    //}
    fmt.Printf("%s %s   %-11s%-8s%-8s%-20s%-15s%s\n",gcheck[y].programName,dotty,pid,gcheck[y].cpuUsage,gcheck[y].memoryUsage,gcheck[y].startTime,gcheck[y].elapseTime,status) 
  }
 
  for y=0;y<configLen;y++ {
    if( config[y].status == 0 ) {
       status=fmt.Sprintf("%s[ KO ]%s",RED,WHITE)
       tmp=len(config[y].programName)
       nbdot=progNameLen-tmp
       dotty = fmt.Sprintf("%*s",nbdot," ")
       dotty = strings.Replace(dotty," ",".",-1)
       fmt.Printf("%s %s   %-11s%-8s%-8s%-20s%-15s%s\n",config[y].programName,dotty,"-","-","-","-","-",status) 
    }
    config[y].status = 0
  }
 
}
 
func getUptime()(float64) {
 
  //var uptime uint64 
  var uptime float64
  uptimef := fmt.Sprintf("/proc/uptime")
  c, err := os.Open(uptimef)
  if err != nil {
    return 1
  }
  defer c.Close()
  scanner := bufio.NewScanner(c)
  for scanner.Scan() {
    stats := strings.Split(scanner.Text()," ")
    //uptime,_ = strconv.Atoi(stats[1])
    //i,_ := strconv.ParseUint(stats[1],10,64)
    uptime,_ = strconv.ParseFloat(stats[0], 64)
  }
  return float64(uptime)
}
 
func getMemTotal()(int) {
 
  var regexpr *regexp.Regexp
  c,err := os.Open("/proc/meminfo")
  if err != nil {
    return -1
  }
  defer c.Close()
  regexpr = regexp.MustCompile("MemTotal")
  scanner := bufio.NewScanner(c)
  for scanner.Scan() {
    stats := strings.Split(strings.Replace(strings.Replace(scanner.Text(),"kB","",-1)," ","",-1),":")
    if ( regexpr.MatchString(stats[0]) ) {
      MEM_TOTAL,_ = strconv.ParseUint(stats[1],10, 64)
    }
  }
  return 0
}
 
func memoryUsage(pid int) (float64) {
 
  var (
     vmsize, rsssize, page_size,memtotal uint64
     memory_usage float64
  )
 
  page_size = PAGE_SIZE
  memtotal = MEM_TOTAL
  statf := fmt.Sprintf("/proc/%d/statm", pid)
  c, err := os.Open(statf)
  if err != nil {
    return 1
  }
  defer c.Close()
  scanner := bufio.NewScanner(c)
  for scanner.Scan() {
    stats := strings.Split(scanner.Text()," ")
    rsssize,_ = strconv.ParseUint(stats[1],10, 64) 
  }
  vmsize = rsssize * page_size / 1024
  memory_usage = 100 * ( float64(vmsize) / float64(memtotal))
  return memory_usage
}
 
func cpuUsage(pid int) (float64) {
  
  var (
    //utime,stime,cutime,cstime,starttime,uptime,sec,cpu_usage,total_time,seconds uint64
    utime,stime,starttime,total_time,hertz uint64
    uptime,cpu_usage,seconds float64
  )
  hertz = CLOCK_TICK
  uptime = getUptime()
  statf := fmt.Sprintf("/proc/%d/stat", pid)
  c, err := os.Open(statf)
  if err != nil {
    return 1
  }
  defer c.Close()
 
  scanner := bufio.NewScanner(c)
  for scanner.Scan() {
    stats := strings.Split(scanner.Text()," ")
    utime,_ = strconv.ParseUint(stats[13],10, 64)
    stime,_ = strconv.ParseUint(stats[14],10, 64)
    //cutime,_ = strconv.ParseUint(stats[16],10, 64)
    //cstime,_ = strconv.ParseUint(stats[17],10, 64)
    starttime,_ = strconv.ParseUint(stats[21],10, 64)
  }
  //fmt.Printf("utime:%d , stime:%d, uptime:%f, starttime:%d, hertz:%d\n",utime,stime,uptime,starttime,hertz)
  total_time = utime + stime
  seconds = uptime - (float64(starttime)/float64(hertz))
  cpu_usage=100*((float64(total_time)/float64(hertz))/seconds) 
  //fmt.Printf("CPU USAGE : %f\n",cpu_usage)
//  fmt.Println(cpu_usage)
 
  return cpu_usage
}
 
// search for the process from the config file
func findProcess(pid int) (int, error) {
 
  var(
   regexpr *regexp.Regexp
   param,configLen,i int
  )
 
  configLen = confCnt
 
  dir := fmt.Sprintf("/proc/%d", pid)
  cmdline := fmt.Sprintf("%s/cmdline", dir)
  _, err := os.Stat(dir)
  if err != nil {
    if os.IsNotExist(err) {
      return 1, nil
    }
    return 1, nil
  }
  
  c, err := os.Open(cmdline)
  if err != nil {
    return 1,err  
  } 
  defer c.Close()
 
  cmdstr := make([]byte,1000)
  cmdsize,err := c.Read(cmdstr)
  if err != nil {
    return 1,err
  }
 
  cmdarr := strings.Split(string(cmdstr[:cmdsize]),"\x00")
  //fmt.Println(cmdarr)
  for i=0;i<configLen;i++ {
    regexpr = regexp.MustCompile(config[i].processString) 
    if ( regexpr.MatchString(string(cmdstr[:cmdsize])) ) {
      gcheck[procCnt].programName = fmt.Sprintf("%s",config[i].programName)
      gcheck[procCnt].pid = pid
      gcheck[procCnt].cpuUsage = fmt.Sprintf("%3.2f",cpuUsage(pid))
      gcheck[procCnt].memoryUsage = fmt.Sprintf("%3.2f",memoryUsage(pid))
      param,_ = strconv.Atoi(config[i].paramToAdd)
      if ( param > 0 ) {
        gcheck[procCnt].programName = fmt.Sprintf("%s:%s",config[i].programName,cmdarr[param])
      }
      gcheck[procCnt].status = 1
      dinfo,_ := os.Stat(dir)
      stat_t := dinfo.Sys().(*syscall.Stat_t)
      stime := time.Unix(int64(stat_t.Ctim.Sec),0)
      gcheck[procCnt].startTime = fmt.Sprintf("%s", stime.Format("2006-01-02 15:04"))
      gcheck[procCnt].elapseTime = convertTimestamp(uint64(stat_t.Ctim.Sec))
      config[i].status = 1
      procCnt++
    }
  }
  return 0,nil
}
 
// loop over /proc/pid
func getProcess()(int, error) {
 
  proc_dir, err := os.Open("/proc")
  if err != nil {
    return 1,err
  }
  defer proc_dir.Close()
  
  for {
  
    pid_dir, err := proc_dir.Readdir(10)
    if err == io.EOF {
	break
    }
    if err != nil {
      return 1,err
    } 
    
    for _, process := range pid_dir {
    
      pid,_ := strconv.Atoi(process.Name())
       
      if pid > 0 {
        findProcess(pid)
      }
    
    }
  }
  return 0,nil
}
 
var Usage = func () {
  fmt.Printf("Usage of %s:\n",os.Args[0])
  flag.PrintDefaults()
  
  fmt.Printf("\n%s -config <configfile>\n",os.Args[0])
  fmt.Printf("\nAuthor: RAMJANALLY Ghoulseine\n")
  fmt.Printf("Version: 0.1\n")
}
 
 
func main() {
 
 var (
   configFile string
   i,repeat,interval int
 )
 repeat=0
 interval=1 
 flag.StringVar(&configFile,"config", "","configuration file, mandatory option nothing will start without it !") 
 flag.IntVar(&repeat,"repeat", 0,"repeat n times this command (default 0)") 
 flag.IntVar(&interval,"interval", 1,"interval factor time per second, to be used with repeat") 
 flag.Parse() 
 
 if len(os.Args) == 1 {
    Usage()
    os.Exit(1)
 }
 
 if configFile != "" {
   readConfig(configFile)
   getClockTicksPerSecond()
   getMemTotal()
   getPageSize()
 
   if(repeat == 0 ) {
   getProcess()
   printProcess()
   }else{
     for i=0;i<repeat;i++ {
       getProcess()
       printProcess()
       procCnt=0
       fmt.Printf("\n")
       time.Sleep( time.Duration(interval) * time.Second )
     }
   }
 }
 
}
