package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	//netSource = "!it2"
	netSource = "pr$"
	netAddr   = "10.32.1.23"
)

var (
	//Срез для записи действий в лог
	mainLog []string

	//netAddr = os.Getenv("SystemDrive")
)

func setMainLog(s string) {
	mainLog = append(mainLog, s)
}
func scanDir(path string) (map[string]time.Time, error) {
	listFiles := make(map[string]time.Time)
	a := make(map[string]time.Time)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		filePath := filepath.Join(path, file.Name())
		if file.IsDir() {
			a, err = scanDir(filePath)
			for k, v := range a {
				listFiles[k] = v
			}
			if err != nil {
				return nil, err
			}
		}
		listFiles[filePath] = file.ModTime()
	}
	return listFiles, nil
}
func validateLenDir(allScans map[string]time.Time, path string) map[string]time.Time {
	var p = make([]string, len(allScans))
	tmpBool := false
	for key, _ := range allScans {
		if len([]rune(key)) > 250 {
			tmpBool = true
			p = strings.Split(key, "\\")
			p1 := strings.Split(key, "\\")
			var pathTmp, p1st string
			for i := len(p) - 1; i >= 0; i-- {
				p1st = strings.Join(p1[:i+1], "\\")
				if len([]rune(strings.Join(p, "\\"))) > 250 {
					if len([]rune(p[i])) > 40 {
						p[i] = string([]rune(p[i])[:10]) + "((CAT))" + string([]rune(p[i])[len([]rune(p[i]))-14:])
						pathTmp = strings.Join(p[:i+1], "\\")
						strLog := fmt.Sprintf("МЕНЯЮ ИЗ: %s\nВ: %s\n", p1st, pathTmp)
						setMainLog(strLog)
						_ = os.Rename(p1st, pathTmp)
					}
				}
			}
		}
	}
	if tmpBool {
		allScans, _ = scanDir(path)
	}
	return allScans
}
func isValidDir(netDir string) bool {
	r := false
	_, err := os.Open(netDir + "\\")
	if err == nil {
		r = true
	}
	return r
}
func isNoValidDirMakeDir(s string) {
	sl := strings.Split(s, "\\")
	for i, _ := range sl {
		if i < len(sl) {
			stt := strings.Join(sl[:i+1], "\\")
			if !isValidDir(stt) {
				if stt != "\\" || stt != "\\\\"+netAddr {
					setMainLog("Создаю директорию: " + stt + "\n")
					err := os.Mkdir(stt, 0700)
					if err != nil {
						setMainLog("Не удалось создать директорию: " + stt + "\n")
					}
				}
			}
		}
	}
}
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
func saveInTrustPlacies(m []string, i int, s string) (int, error) {
	var count int
	for j, path := range m {
		file, err := os.Stat(path)
		if err != nil {
			return j, err
		}
		p := strings.Split(path, "\\")
		partPath := p[i:]
		if file.IsDir() {
			if !isValidDir(s + strings.Join(partPath, "\\")) {
				strLog := fmt.Sprintf("\nСоздаю директорию: %v\n", s+strings.Join(partPath, "\\"))
				setMainLog(strLog)
				err := os.Mkdir(s+strings.Join(partPath, "\\"), 0700)
				if err != nil {
					return j, err
				}
			}
			continue
		} else {
			strLog := fmt.Sprintf("Копирую файл: %v\n в %v\n", path, s+strings.Join(partPath, "\\"))
			setMainLog(strLog)
			err := Copy(path, s+strings.Join(partPath, "\\"))
			if err != nil {
				return j, err
			}
		}
		count++
	}
	return count, nil
}
func getSliceFromMapKeys(m map[string]time.Time, i int) ([]string, []string) {
	isNoZero := false
	keys := make([]string, 0, len(m))
	delPath := make([]string, 0, len(m))
	for k := range m {
		key := strings.Split(k, "\\")[i:]
		if i != 0 {
			delPath = append(delPath, strings.Join(strings.Split(k, "\\")[:i], "\\"))
			isNoZero = true
		}
		keys = append(keys, strings.Join(key, "\\"))
	}
	sort.Strings(keys)
	if isNoZero == true {
		return keys, delPath
	} else {
		return keys, nil
	}
}
func mainAction(usrProf string, netPath string, value string) {
	setMainLog("***ДИРЕКТОРИЯ ::" + value + ":: ***")

	//Количество файлов на сетевом ресурсе, перемещенных в папку Deleted - на локальном их нет
	var countDel, count int

	//Создаю папки, в которые копируются папки поумолчанию и копирую данные
	//Если нужные папки из поумолчанию не существуют на сетевом ресурсе, то создаю их и копирую данные
	if !isValidDir(netPath + "\\" + value) {
		err := os.Mkdir(netPath+"\\"+value, 0700)
		if err != nil {
			setMainLog("Не удалось создать папку: " + netPath + "\\" + value + "\n")
		}
		setMainLog("Директория создана - " + netPath + "\\" + value + "\n")
		copyListNess, err := scanDir(usrProf + "\\" + value)
		if err != nil {
			strLog := fmt.Sprintf("Ошибка при сканировании директории %s ::: %v\n", value, err)
			setMainLog(strLog)
		}
		copyListNess = validateLenDir(copyListNess, usrProf+"\\"+value)
		copyListNesSlice, _ := getSliceFromMapKeys(copyListNess, 0)
		count, err := saveInTrustPlacies(copyListNesSlice, len(strings.Split(usrProf, "\\")), netPath+"\\")
		if err != nil {
			strLog := fmt.Sprintf("%v\n", err)
			setMainLog(strLog)
		}
		setMainLog("В " + value + " скопировано объектов: " + strconv.Itoa(count) + "\n")
	} else {
		//Папки на сервере существуют, значит надо сравнить источник и приемник
		setMainLog("Сравниваю даты модификации в локальном расположении: " + usrProf + "\\" + value + " с датами в сетевом расположении\n")
		localSourceCopy, err := scanDir(usrProf + "\\" + value)
		if err != nil {
			strLog := fmt.Sprintf("%v\n", err)
			setMainLog(strLog)
		}
		localSourceCopy = validateLenDir(localSourceCopy, usrProf+"\\"+value)
		localSourceCopySlice, _ := getSliceFromMapKeys(localSourceCopy, len(strings.Split(usrProf, "\\")))
		netSourceCopy, err := scanDir(netPath + "\\" + value)
		if err != nil {
			strLog := fmt.Sprintf("%v\n", err)
			setMainLog(strLog)
		}
		netSourceCopy = validateLenDir(netSourceCopy, netPath+"\\"+value)
		netSourceCopySlice, _ := getSliceFromMapKeys(netSourceCopy, len(strings.Split(netPath, "\\")))

		//Проверяю, есть ли на ресурсе папка Deleted, если нет - создаю ее
		if !isValidDir(netPath + "\\Deleted") {
			err := os.Mkdir(netPath+"\\Deleted", 0700)
			if err != nil {
				setMainLog("Не удалось создать папку: " + netPath)
				os.Exit(1)
			}
			setMainLog("Директория создана - " + netPath + "\\Deleted")
		}
		var countCopy int
		for _, localSlice := range localSourceCopySlice {
			var singleFile []string
			isDuplicate := false
			i := sort.Search(len(netSourceCopySlice), func(i int) bool { return netSourceCopySlice[i] >= localSlice })
			if i < len(netSourceCopySlice) && netSourceCopySlice[i] == localSlice {
				isDuplicate = true
			}
			/*
				Сильно увеличилась скорость сравнения ресурсов, перейдя с линейного поиска - ниже
				на бинарный поиск - выше.
				for _, netSlice := range netSourceCopySlice {
					if localSlice == netSlice {
						isDuplicate = true
						continue
					}
				}
			*/
			if isDuplicate {
				//Если есть одинаковые файлы, сравниваю их даты модификации
				if localSourceCopy[usrProf+"\\"+localSlice].After(netSourceCopy[netPath+"\\"+localSlice]) {
					//fmt.Println(strings.Split(usrProf+"\\"+localSlice, "\\"))
					singleFile = append(singleFile, usrProf+"\\"+localSlice)
					count, err = saveInTrustPlacies(singleFile, len(strings.Split(usrProf, "\\")), netPath+"\\")
					if err != nil {
						strLog := fmt.Sprintf("%v\n", err)
						setMainLog(strLog)
					}
				}
			} else {
				//Записать файлы(папки) которых нет в сетевом архиве
				singleFile = append(singleFile, usrProf+"\\"+localSlice)
				countCopy, err = saveInTrustPlacies(singleFile, len(strings.Split(usrProf, "\\")), netPath+"\\")
				if err != nil {
					strLog := fmt.Sprintf("%v\n", err)
					setMainLog(strLog)
				}
			}
		}
		if countCopy > 0 {
			setMainLog("В " + value + " добавлено " + strconv.Itoa(countCopy) + " файлов\n")
		}
		if count > 0 {
			setMainLog("В " + value + " произведена замена " + strconv.Itoa(count) + " файлов\n")
		}
		//Часть для работы на удаленном ресурсе с файлами, которые были удалены на локальном
		for _, netSlice := range netSourceCopySlice {
			var singleFile []string
			isDuplicate := false
			i := sort.Search(len(localSourceCopySlice), func(i int) bool { return localSourceCopySlice[i] >= netSlice })
			if i < len(localSourceCopySlice) && localSourceCopySlice[i] == netSlice {
				isDuplicate = true
			}
			/*
				Сильно увеличилась скорость сравнения ресурсов, перейдя с линейного поиска - ниже
				на бинарный поиск - выше.
				for _, localSlice := range localSourceCopySlice {
					if netSlice == localSlice {
						isDuplicate = true
					}
				}
			*/
			if !isDuplicate {
				//Надо создать папки пути
				isNoValidDirMakeDir(netPath + "\\Deleted\\" + netSlice)
				singleFile = append(singleFile, netPath+"\\"+netSlice)
				strLog := fmt.Sprintf("Перемещаю файл: %v\nв %v\n", netSlice, netPath+"\\Deleted\\"+netSlice)
				setMainLog(strLog)
				countDel, err = saveInTrustPlacies(singleFile, len(strings.Split(netPath, "\\")), netPath+"\\Deleted\\")
				if err != nil {
					strLog := fmt.Sprintf("%v\n", err)
					setMainLog(strLog)
				}
				err = os.Remove(netPath + "\\" + netSlice)
			}
		}
		if countDel > 0 {
			setMainLog("\nВ папку Deleted перемещено " + strconv.Itoa(countDel) + " файлов")
		}
	}
}

func main() {
	var (
		adUser   = os.Getenv("USERNAME")
		compName = os.Getenv("COMPUTERNAME")
		//netPath  = netAddr + "\\" + netSource + "\\" + adUser + "\\" + compName
		netPath = "\\\\" + netAddr + "\\" + netSource + "\\" + adUser + "\\" + compName
		usrProf = os.Getenv("USERPROFILE")
		//Папки, которые копируются поумолчанию
		netDirs = []string{"Desktop", "Documents", "Downloads", "Favorites"}
		//Папки, заданные в аргументах коммандной строки
		mailStrings, otherStrings []string
	)
	/* Блок аргументов командной строки*/
	if len(os.Args) > 1 {
		if len(os.Args) == 2 && (os.Args[1] == "/?" || os.Args[1] == "-h" || os.Args[1] == "--help") {
			fmt.Printf("\nHow to use this application:\n\n"+
				"Default directories for make backup is: 'Desktop, Documents, Downloads, Favorites'\n\n"+
				"Also, you can use command-line prompts.\n\n"+
				"Key (prompt) for copying directories of email client:\n"+
				"\tmail:live - for copying directories of Windows Live Mail\n"+
				"\tmail:outlook - for copying directories of MS Outlook\n\n"+
				"For copying of no-default directories,- write full-path without any keys:\n"+
				"\tc:\\123\\my_files\n"+
				"or \t\"c:\\my files\"\n\n"+
				"If path has a spase, then all the path must be in quotes.\n"+
				"For example:\n\t%s mail:live  mail:outlook d:\\all_files\\for_me \"c:\\dads files\"\n\n", filepath.Base(os.Args[0]))
			os.Exit(1)
		} else {
			for i := 1; i < len(os.Args); i++ {
				if strings.HasPrefix(os.Args[i], "mail") {
					switch pp := os.Args[i]; pp {
					case "mail:live":
						if isValidDir(filepath.Join(usrProf, "AppData\\Local\\Microsoft\\Windows Live Mail")) {
							mailStrings = append(mailStrings, "AppData\\Local\\Microsoft\\Windows Live Mail")
						}
					case "mail:outlook":
						if isValidDir(filepath.Join(usrProf, "AppData\\Local\\Microsoft\\Outlook")) {
							mailStrings = append(mailStrings, "AppData\\Local\\Microsoft\\Outlook")
						}
					}
				}
				if isValidDir(strings.TrimSpace(os.Args[i])) {
					otherStrings = append(otherStrings, strings.TrimSpace(os.Args[i]))
				}
			}
		}
	}
	timeStart := time.Now()
	fmt.Println(time.Unix(timeStart.Unix(), 0))
	fmt.Println("Запущено резервное копирование")
	/*Проверяю, готов ли сетевой ресурс - ЛОКАЛЬНЫЙ ТЕСТ
	if !isValidDir(netAddr + "\\" + netSource) {
		fmt.Println("Нет такого ресурса", netAddr+"\\"+netSource, "Создайте эту шару и запустите копирование заново")
		os.Exit(1)
	}
	*/

	/*Проверяю наличие сетевого ресурса::: РАБОЧАЯ ВЕРСИЯ*/
	if !isValidDir("\\\\" + netAddr + "\\" + netSource) {
		fmt.Println("Нет такого ресурса \\\\"+netAddr+"\\"+netSource, "Создайте эту шару и запустите копирование заново")
		os.Exit(1)
	}

	//Проверяю, есть ли на ресурсе папка ДОМенноеимя/имя_компьютера, если нет - создаю ее
	isNoValidDirMakeDir(netPath)

	//Отправляю в горутинах папки для копирования/замены
	for _, value := range netDirs {
		mainAction(usrProf, netPath, value)
	}

	//Для папок из аргументов командной строки
	if len(os.Args) > 1 {
		//Для заданных папок
		var (
			usr string
			net string
		)

		//Для почты, указанной ключами
		for _, val := range mailStrings {
			isNoValidDirMakeDir(netPath + "\\" + val)
			mainAction(usrProf, netPath, val)
		}

		//Для аргументов-непочта
		for _, value := range otherStrings {
			usr = strings.Split(value, "\\")[0]
			isNoValidDirMakeDir(netPath + "\\" + strings.Join(strings.Split(value, "\\")[1:], "\\"))
			net = strings.Join(strings.Split(value, "\\")[1:], "\\")
			mainAction(usr, netPath, net)
		}
	}
	file, _ := os.Create(netPath + "\\old_log.txt")
	timeEnd := time.Now()
	strLog := fmt.Sprintf("time start - %v\ntime end %v\n", timeStart, timeEnd)
	setMainLog(strLog)
	defer file.Close()
	for i := range mainLog {
		file.WriteString("\r\n" + mainLog[i] + "\r\n")
	}
}
