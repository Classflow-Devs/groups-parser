package dto

type GroupDTO struct {
	GroupName          string `json:"groupName"`
	GroupID            int    `json:"groupID"`
	Course             int    `json:"course"`
	FormID             int    `json:"formID"`
	FormStud           string `json:"formStud"`
	YearName           string `json:"yearName"`
	FacultyID          int    `json:"facultyID"`
	PlanID             *int   `json:"planID"`
	LevelID            int    `json:"levelID"`
	SpecialtyIDAndName string `json:"specialtyIDAndName"`
	SpecialtyID        int    `json:"specialtyID"`
}

type GroupResponseDTO struct {
	Data struct {
		Groups              []GroupDTO `json:"groups"`
		ConditionsEducation []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"conditionsEducation"`
		LevelEducation []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"levelEducation"`
		Nationality []string `json:"nationality"`
		Specialties []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"specialties"`
		StatusStudents []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"statusStudents"`
	} `json:"data"`
	State int     `json:"state"`
	Msg   string  `json:"msg"`
	Time  float64 `json:"time"`
}

type GroupInfoDTO struct {
	Data struct {
		FaculName        string `json:"faculName"`
		KafName          string `json:"kafName"`
		FormName         string `json:"formName"`
		LevelName        string `json:"levelName"`
		GroupYear        string `json:"groupYear"`
		StudentInfoGroup []struct {
			StudentID     int         `json:"studentID"`
			FullName      string      `json:"fullName"`
			NumRecordBook string      `json:"numRecordBook"`
			PhotoLink     string      `json:"photoLink"`
			NumberMobile  string      `json:"numberMobile"`
			Email         string      `json:"email"`
			AdmissionYear string      `json:"admissionYear"`
			IsLocked      interface{} `json:"isLocked"`
			IsLockedVed   bool        `json:"isLockedVed"`
			JournalData   interface{} `json:"journalData"`
			YearList      interface{} `json:"yearList"`
			Online        bool        `json:"online"`
			Rating        interface{} `json:"rating"`
			LastEnterDate interface{} `json:"lastEnterDate"`
		} `json:"studentInfoGroup"`
		Group struct {
			Item1 string `json:"item1"`
			Item2 int    `json:"item2"`
		} `json:"group"`
		Course      string `json:"course"`
		SpecialName string `json:"specialName"`
		Op          string `json:"op"`
		Plan        struct {
			Item1 interface{} `json:"item1"`
			Item2 int         `json:"item2"`
		} `json:"plan"`
		TrainingDirection string `json:"trainingDirection"`
		SemList           []struct {
			Value int    `json:"value"`
			Text  string `json:"text"`
		} `json:"semList"`
		AllowChangePass bool        `json:"allowChangePass"`
		ShowRaspButton  bool        `json:"showRaspButton"`
		LinkRaspButton  interface{} `json:"linkRaspButton"`
		ShowGraphButton bool        `json:"showGraphButton"`
		ShowVedButton   bool        `json:"showVedButton"`
		HideStudents    bool        `json:"hideStudents"`
		Hedden          bool        `json:"hedden"`
		LabelFaculName  string      `json:"labelFaculName"`
		CuratorUserID   interface{} `json:"curatorUserID"`
		CuratorUserName interface{} `json:"curatorUserName"`
	} `json:"data"`
	State int     `json:"state"`
	Msg   string  `json:"msg"`
	Time  float64 `json:"time"`
}
