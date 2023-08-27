package bcutasr

const (
	API_REQ_UPLOAD    = "https://member.bilibili.com/x/bcut/rubick-interface/resource/create"
	API_COMMIT_UPLOAD = "https://member.bilibili.com/x/bcut/rubick-interface/resource/create/complete"
	API_CREATE_TASK   = "https://member.bilibili.com/x/bcut/rubick-interface/task"
	API_QUERY_RESULT  = "https://member.bilibili.com/x/bcut/rubick-interface/task/result"
)

var (
	SUPPORT_SOUND_FORMAT = []string{"flac", "aac", "m4a", "mp3", "wav"}
)

const (
	ResultStateStop = iota
	ResultStateRuning
	_
	ResultStateError
	ResultStateComplete
)
