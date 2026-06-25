package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Date      string `json:"date"`
	Completed bool   `json:"completed"`
	Marked    bool   `json:"marked"`
	IsDeleted bool   `json:"isDeleted"`
}

type AppConfig struct {
	Theme         string `json:"theme"`
	Lang          string `json:"lang"`
	FocusSubtitle string `json:"focusSubtitle"`
}

type InitResponse struct {
	Tasks  []Task    `json:"tasks"`
	Config AppConfig `json:"config"`
}

type CommonResponse struct {
	Status string `json:"status"`
	Tasks  []Task `json:"tasks,omitempty"`
}

var db *sql.DB
var dataDir string

func main() {
	// 1. 动态获取绿联持久化数据路径
	dataDir = os.Getenv("PERSISTENT_DATA_DIR")
	if dataDir == "" {
		dataDir = "./database"
	}
	_ = os.MkdirAll(dataDir, 0755)

	dbPath := filepath.Join(dataDir, "will_station.db")
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("DB open failed: %v", err)
	}
	defer db.Close()

	initDatabaseTables()

	// 2. 🌟【终极修复】绝对路径解包定位：直接抓取二进制程序平级的 www 目录
	exePath, err := os.Executable()
	if err != nil {
		exePath = "."
	}
	exeDir := filepath.Dir(exePath)
	webRoot := filepath.Join(exeDir, "www")

	log.Printf("[UGOS Pro Native] Executable: %s, WebRoot locked at: %s", exePath, webRoot)
	fileServer := http.FileServer(http.Dir(webRoot))

	// 3. 健康检查与 API/静态资源分流网闸
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 🌟 100% 放行绿联测活探针对 / 或 /health 的探测，无脑回传 200，破除转圈魔咒！
		if r.URL.Path == "/" || r.URL.Path == "/health" {
			if _, err := os.Stat(filepath.Join(webRoot, "index.html")); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"UP","msg":"Will ToDo Core Activating"}`))
				return
			}
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	// 注册核心交互路由
	http.HandleFunc("/api/init", corsMiddleware(handleInit))
	http.HandleFunc("/api/config", corsMiddleware(handleSaveConfig))
	http.HandleFunc("/api/tasks", corsMiddleware(handleAddTask))
	http.HandleFunc("/api/tasks/update", corsMiddleware(handleUpdateTask))
	http.HandleFunc("/api/tasks/delete", corsMiddleware(handlePermanentDeleteTask))
	http.HandleFunc("/api/tasks/reset", corsMiddleware(handleResetAllData))
	http.HandleFunc("/api/tasks/export", corsMiddleware(handleExportCSV))

	// 4. 完美锁死 project.yaml 声明的 3177 专属隔离端口
	port := os.Getenv("APPS_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "3177"
	}

	log.Printf("[UGOS Pro Success] Service listening on port: :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server start failed: %v", err)
	}
}

func initDatabaseTables() {
	taskTable := `CREATE TABLE IF NOT EXISTS tasks (id TEXT PRIMARY KEY, text TEXT, date TEXT, completed INTEGER DEFAULT 0, marked INTEGER DEFAULT 0, is_deleted INTEGER DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);`
	configTable := `CREATE TABLE IF NOT EXISTS app_config (id INTEGER PRIMARY KEY CHECK (id = 1), theme TEXT DEFAULT 'light', lang TEXT DEFAULT 'zh', focus_subtitle TEXT DEFAULT '');`
	_, _ = db.Exec(taskTable)
	_, _ = db.Exec(configTable)
	_, _ = db.Exec("INSERT OR IGNORE INTO app_config (id, theme, lang, focus_subtitle) VALUES (1, 'light', 'zh', '')")
}
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}
func getAllTasksFromDB() ([]Task, error) {
	rows, err := db.Query("SELECT id, text, date, completed, marked, is_deleted FROM tasks ORDER BY completed ASC, date ASC, created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Task
	for rows.Next() {
		var t Task
		var comp, mark, del int
		if err := rows.Scan(&t.ID, &t.Text, &t.Date, &comp, &mark, &del); err != nil {
			return nil, err
		}
		t.Completed = comp == 1
		t.Marked = mark == 1
		t.IsDeleted = del == 1
		list = append(list, t)
	}
	if list == nil {
		list = []Task{}
	}
	return list, nil
}
func handleInit(w http.ResponseWriter, r *http.Request) {
	tasks, _ := getAllTasksFromDB()
	var conf AppConfig
	_ = db.QueryRow("SELECT theme, lang, focus_subtitle FROM app_config WHERE id = 1").Scan(&conf.Theme, &conf.Lang, &conf.FocusSubtitle)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(InitResponse{Tasks: tasks, Config: conf})
}
func handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var conf AppConfig
	if err := json.NewDecoder(r.Body).Decode(&conf); err == nil {
		_, _ = db.Exec("UPDATE app_config SET theme = ?, lang = ?, focus_subtitle = ? WHERE id = 1", conf.Theme, conf.Lang, conf.FocusSubtitle)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CommonResponse{Status: "success"})
}
func handleAddTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string
		Date string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		_, _ = db.Exec("INSERT INTO tasks (id, text, date, completed, marked, is_deleted) VALUES (?, ?, ?, 0, 0, 0)", id, req.Text, req.Date)
	}
	updatedTasks, _ := getAllTasksFromDB()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CommonResponse{Status: "success", Tasks: updatedTasks})
}
func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
		for k, v := range body {
			switch k {
			case "date":
				_, _ = db.Exec("UPDATE tasks SET date = ? WHERE id = ?", v.(string), id)
			case "completed":
				val := 0
				if v.(bool) {
					val = 1
				}
				_, _ = db.Exec("UPDATE tasks SET completed = ? WHERE id = ?", val, id)
			case "marked":
				val := 0
				if v.(bool) {
					val = 1
				}
				_, _ = db.Exec("UPDATE tasks SET marked = ? WHERE id = ?", val, id)
			case "isDeleted":
				val := 0
				if v.(bool) {
					val = 1
				}
				_, _ = db.Exec("UPDATE tasks SET is_deleted = ? WHERE id = ?", val, id)
			}
		}
	}
	updatedTasks, _ := getAllTasksFromDB()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CommonResponse{Status: "success", Tasks: updatedTasks})
}
func handlePermanentDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	_, _ = db.Exec("DELETE FROM tasks WHERE id = ?", id)
	updatedTasks, _ := getAllTasksFromDB()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CommonResponse{Status: "success", Tasks: updatedTasks})
}
func handleResetAllData(w http.ResponseWriter, r *http.Request) {
	_, _ = db.Exec("DELETE FROM tasks")
	_, _ = db.Exec("UPDATE app_config SET theme='light', lang='zh', focus_subtitle='' WHERE id=1")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CommonResponse{Status: "success"})
}
func handleExportCSV(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	rows, err := db.Query("SELECT text, date, completed, marked FROM tasks WHERE is_deleted = 0 ORDER BY date ASC")
	if err != nil {
		return
	}
	defer rows.Close()
	var csvBuilder strings.Builder
	csvBuilder.Write([]byte{0xEF, 0xBB, 0xBF})
	if lang == "zh" {
		csvBuilder.WriteString("到期日,事项,标记,状态\n")
	} else {
		csvBuilder.WriteString("Due Date,Task,Marked,Status\n")
	}
	for rows.Next() {
		var text, date string
		var comp, mark int
		if err := rows.Scan(&text, &date, &comp, &mark); err == nil {
			safeText := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(text, "\n", " "), "\r", " "), "\"", "\"\"")
			mStr := "/"
			if mark == 1 {
				if lang == "zh" {
					mStr = "已标记"
				} else {
					mStr = "Marked"
				}
			}
			sStr := "未完成"
			if comp == 1 {
				if lang == "zh" {
					sStr = "已完成"
				} else {
					sStr = "Done"
				}
			} else {
				if lang != "zh" {
					sStr = "Pending"
				}
			}
			csvBuilder.WriteString(fmt.Sprintf("%s,\"%s\",%s,%s\n", date, safeText, mStr, sStr))
		}
	}
	w.Header().Set("Content-Disposition", "attachment; filename=will_todo_backup.csv")
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvBuilder.String()))
}
