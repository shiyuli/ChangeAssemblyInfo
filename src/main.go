package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	versionFileName string = "AssemblyInfo.cs"
	versionStoreFileName string = "version.txt"
)

/*
Usage: ChangeAssemblyInfo.exe ./AssemblyInfo.cs 0.0.0.0
 */
func main() {
	fmt.Println("Changing assembly version ...")
	defer fmt.Println("Done!")

	if len(os.Args) < 2 {
		fmt.Println("The number of arguments can not less than 2!")
		return
	}

	var filename string = os.Args[1]
	fmt.Println("Target:", filename)

	if !checkFilename(filename) {
		fmt.Println("first command line argument is not a valid filename")
		return
	}

	version, err := getVersion()
	if err != nil {
		fmt.Println(err)
		return
	}

	changeVersion(filename, version)
	fmt.Println("Assembly version changed to:", version)
}

func checkFilename(filename string) bool {
	var filenameArray []string = strings.Split(filename, "\\")
	return filenameArray[len(filenameArray) - 1] == versionFileName
}

func changeVersion(filename string, version string) {
	fileContentArray, err := readFile(filename)
	if err != nil {
		return
	}

	for index, line := range fileContentArray {
		if isVersionField(line) {
			replaceVersion(&fileContentArray[index], version)
		}
	}

	writeFile(filename, fileContentArray)
}

func readFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var result []string
	iReader := bufio.NewReader(file)
	for {
		rawLine, _, err := iReader.ReadLine()
		var strLine string = string(rawLine)

		if err == nil {
			result = append(result, strLine)
		} else if err == io.EOF {
			return result, nil
		} else {
			return nil, err
		}
	}
}

func writeFile(filename string, content []string) {
	outputFile, err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}

	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)
	outputString := strings.Join(content, "\n")

	outputWriter.WriteString(outputString)
	outputWriter.Flush()
}

func isVersionField(input string) bool {
	//[assembly: AssemblyVersion("*.*.*.*")]
	//[assembly: AssemblyFileVersion("*.*.*")]

	if len(input) >= 2 && input[:2] == "//" {
		return false
	}

	if strings.Contains(input, "AssemblyVersion") || strings.Contains(input, "AssemblyFileVersion") {
		return true
	} else {
		return false
	}
}

func getVersion() (string, error) {
	//versionStoreFile struct:
	//BigVersion
	//MonthlyBuild
	//DailyBuild
	//CIBuild

	// use third command line argument as version instead of versionStoreFile
	if len(os.Args) > 2 {
		version := os.Args[2]
		if isValidVersion(version) {
			return version, nil
		} else {
			return "", errors.New("third command line argument is not a valid version number")
		}
	}

	content, err := readFile(versionStoreFileName)

	if err == nil && len(content) > 0 {
		ciVersion, _ := strconv.Atoi(content[3]) //get CIBuild version
		ciVersion++
		content[3] = strconv.Itoa(ciVersion)
		defer writeFile(versionStoreFileName, content)
		version := content[0] + "." + content[1] + "." + content[2] + "." + content[3]
		return version, nil
	} else {
		return "", err
	}
}

func isValidVersion(version string) bool {
	regex, err := regexp.Compile(`^(\d+\.)(\d+\.)(\d+\.)(\d+)$`)

	if err != nil {
		return false
	}

	return regex.MatchString(version)
}

func replaceVersion(versionField *string, version string) {
	//[assembly: AssemblyVersion("*.*.*.*")]
	//[assembly: AssemblyFileVersion("*.*.*.*")]

	if !strings.Contains(*versionField, "AssemblyVersion") &&
		!strings.Contains(*versionField, "AssemblyFileVersion") {
		return
	}

	beginIndex := strings.Index(*versionField, "(")
	endIndex := strings.Index(*versionField, ")")

	leftStr := (*versionField)[:beginIndex+2]
	rightStr := (*versionField)[endIndex-1:]

	*versionField = leftStr + version + rightStr
}
