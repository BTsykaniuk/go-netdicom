// A sample program for issuing C-STORE or C-FIND to a remote server.
package main

import (
	"crypto/tls"
	"flag"
	"github.com/BTsykaniuk/go-netdicom/dimse"
	"log"

	"github.com/BTsykaniuk/go-dicom"
	"github.com/BTsykaniuk/go-dicom/dicomtag"
	"github.com/BTsykaniuk/go-netdicom"
	"github.com/BTsykaniuk/go-netdicom/sopclass"
)

var (
	serverFlag        = flag.String("server", "178.208.149.80:21113", "host:port of the remote application entity")
	storeFlag         = flag.String("store", "", "If set, issue C-STORE to copy this file to the remote server")
	aeTitleFlag       = flag.String("ae-title", "BINOMIX", "AE title of the client")
	remoteAETitleFlag = flag.String("remote-ae-title", "SSLAIBUSSCP", "AE title of the server")
	findFlag          = flag.Bool("find", false, "Issue a C-FIND.")
	getFlag           = flag.Bool("get", false, "Issue a C-GET.")
	seriesFlag        = flag.String("series", "", "Study series UID to retrieve in C-{FIND,GET}.")
	studyFlag         = flag.String("study", "", "Study instance UID to retrieve in C-{FIND,GET}.")
)

func newServiceUser(sopClasses []string) *netdicom.ServiceUser {
	su, err := netdicom.NewServiceUser(netdicom.ServiceUserParams{
		CalledAETitle:    *remoteAETitleFlag,
		CallingAETitle:   *aeTitleFlag,
		SOPClasses:       sopClasses,
		TransferSyntaxes: []string{"1.2.840.10008.1.2.4.50", "1.2.840.10008.1.2.4.51"},
	})
	if err != nil {
		log.Panic(err)
	}

	cert, err := tls.LoadX509KeyPair("client_keystore.crt", "client_keystore.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	log.Printf("Connecting to %s", *serverFlag)
	su.Connect(*serverFlag, &config)
	return su
}

func cStore(inPath string) {
	su := newServiceUser(sopclass.StorageClasses)
	defer su.Release()
	dataset, err := dicom.ReadDataSetFromFile(inPath, dicom.ReadOptions{})
	if err != nil {
		log.Panicf("%s: %v", inPath, err)
	}

	err = su.CStore(dataset)
	if err != nil {
		log.Println(err)
		log.Panicf("%s: cstore failed: %v", inPath, err)
	}
	log.Printf("C-STORE finished successfully")
}

func generateCFindElements(id string) (netdicom.QRLevel, []*dicom.Element) {
	return netdicom.QRLevelStudy, []*dicom.Element{dicom.MustNewElement(dicomtag.StudyInstanceUID, id)}
	//if *seriesFlag != "" {
	//	return netdicom.QRLevelStudy, []*dicom.Element{dicom.MustNewElement(dicomtag.SeriesInstanceUID, *seriesFlag)}
	//}
	//if *studyFlag != "" {
	//	return netdicom.QRLevelStudy, []*dicom.Element{dicom.MustNewElement(dicomtag.StudyInstanceUID, *studyFlag)}
	//}
	//args := []*dicom.Element{
	//	dicom.MustNewElement(dicomtag.AccessionNumber, "*"),
	//	dicom.MustNewElement(dicomtag.ReferringPhysicianName, "*"),
	//	dicom.MustNewElement(dicomtag.PatientName, "*"),
	//	dicom.MustNewElement(dicomtag.PatientID, "*"),
	//	dicom.MustNewElement(dicomtag.PatientBirthDate, "*"),
	//	dicom.MustNewElement(dicomtag.PatientSex, "*"),
	//	dicom.MustNewElement(dicomtag.StudyID, "1.2.276.0.7230010.3.1.2.1787205428.166.1117461927"),
	//	dicom.MustNewElement(dicomtag.RequestedProcedureDescription, "*"),
	//}

}

func cGet() {
	su := newServiceUser(sopclass.QRMoveClasses)
	defer su.Release()

	path := "/Users/admin/Desktop/projects/binomix/go-netdicom/testdata/reportsi.dcm"
	dataset, err := dicom.ReadDataSetFromFile(path, dicom.ReadOptions{})
	if err != nil {
		log.Println(err)
	}
	finding, err := dicom.FindElementByName(dataset.Elements, "StudyInstanceUID")
	if err != nil {
		log.Println(err)
	}

	id := finding.Value[0].(string)

	qrLevel, args := generateCFindElements(id)
	n := 0

	err = su.CMove(qrLevel, args, "GBMAC0261",
		func(transferSyntaxUID, sopClassUID, sopInstanceUID string, data []byte) dimse.Status {
			log.Println("Here")
			log.Printf("%d: C-GET data; transfersyntax=%v, sopclass=%v, sopinstance=%v data %dB",
				n, transferSyntaxUID, sopClassUID, sopInstanceUID, len(data))
			n++
			return dimse.Success
		})
	log.Printf("C-GET finished: %v", err)
}

func cFind() {
	su := newServiceUser(sopclass.QRFindClasses)
	defer su.Release()

	//path := "/Users/admin/Desktop/projects/binomix/go-netdicom/testdata/reportsi.dcm"
	//dataset, err := dicom.ReadDataSetFromFile(path, dicom.ReadOptions{})
	//if err != nil {
	//	log.Println(err)
	//}
	//
	//finding, err := dicom.FindElementByName(dataset.Elements, "StudyInstanceUID")
	//if err != nil {
	//	log.Println(err)
	//}
	//
	//id := finding.Value[0].(string)
	id := "1.2.40.0.13.1.316397237693838535848679192193205474710"
	//log.Println(id)
	var n int
	qrLevel, args := generateCFindElements(id)
	for result := range su.CFind(qrLevel, args) {
		n++
		if result.Err != nil {
			log.Printf("C-FIND error: %v", result.Err)
			continue
		}
		for _, elem := range result.Elements {
			log.Printf("Got elem: %v", elem.String())
		}
	}
	log.Printf("Get %d elements", n-1)
}

func main() {
	//flag.Parse()
	//if *storeFlag != "" {
	//	cStore(*storeFlag)
	//} else if *findFlag {
	//	cFind()
	//} else if *getFlag {
	//	cGet()
	//} else {
	//	log.Panic("Either -store, -get, or -find must be set")
	//}

	cFind()
	//cGet()
	//cStore(path)
}
