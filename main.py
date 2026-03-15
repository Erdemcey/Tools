import sys
import threading
from PyQt6.QtWidgets import *
from PyQt6.QtCore import *
from PyQt6.QtGui import *
from PyQt6.QtWebEngineWidgets import QWebEngineView

from engine import RequestParser
from request import RequestSender

class CodeHighlighter(QSyntaxHighlighter):
    def __init__(self, parent):
        super().__init__(parent)
        self.highlighting_rules = []
        payload_format = QTextCharFormat()
        payload_format.setBackground(QColor("#4b4b00"))
        payload_format.setForeground(QColor("#ffff00"))
        payload_format.setFontWeight(QFont.Weight.Bold)
        self.highlighting_rules.append((r"§.*?§", payload_format))

    def highlightBlock(self, text):
        for pattern, format in self.highlighting_rules:
            expression = QRegularExpression(pattern)
            it = expression.globalMatch(text)
            while it.hasNext():
                match = it.next()
                self.setFormat(match.capturedStart(), match.capturedLength(), format)

class PentestTool(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("Pentest Tool (Intruder & Repeater)")
        self.setGeometry(100, 100, 1400, 900)
        self.init_ui()

    def init_ui(self):
        central = QWidget()
        self.setCentralWidget(central)
        main_layout = QHBoxLayout(central)

        left_splitter = QSplitter(Qt.Orientation.Vertical)
        upper_splitter = QSplitter(Qt.Orientation.Horizontal)
        
        # --- REQUEST ---
        req_container = QWidget()
        req_vbox = QVBoxLayout(req_container)
        req_header_layout = QHBoxLayout()
        req_header_layout.addWidget(QLabel("<b>Request (İstek)</b>"))
        self.check_https = QCheckBox("Use HTTPS")
        req_header_layout.addWidget(self.check_https)
        btn_mark = QPushButton("Add §")
        btn_mark.clicked.connect(self.mark_payload_selection)
        req_header_layout.addWidget(btn_mark)
        req_vbox.addLayout(req_header_layout)

        self.req_edit = QTextEdit()
        self.req_edit.setStyleSheet("font-family: Consolas; background: #1e1e1e; color: #d4d4d4;")
        self.highlighter = CodeHighlighter(self.req_edit.document())
        req_vbox.addWidget(self.req_edit)
        
        # --- RESPONSE ---
        self.res_tabs = QTabWidget()
        self.res_code = QTextEdit(); self.res_code.setReadOnly(True)
        self.res_headers = QTextEdit(); self.res_headers.setReadOnly(True)
        self.res_web = QWebEngineView()
        self.res_tabs.addTab(self.res_code, "Source")
        self.res_tabs.addTab(self.res_headers, "Headers")
        self.res_tabs.addTab(self.res_web, "Render")
        
        upper_splitter.addWidget(req_container)
        upper_splitter.addWidget(self.res_tabs)
        
        self.results_table = QTableWidget(0, 5)
        self.results_table.setHorizontalHeaderLabels(["ID", "Payload", "Status", "Length", "Time (ms)"])
        self.results_table.horizontalHeader().setSectionResizeMode(QHeaderView.ResizeMode.Stretch)

        left_splitter.addWidget(upper_splitter)
        left_splitter.addWidget(self.results_table)
        main_layout.addWidget(left_splitter, 4)

        # --- SETTINGS PANEL ---
        right_panel = QFrame()
        right_panel.setFixedWidth(300)
        right_vbox = QVBoxLayout(right_panel)
        
        right_vbox.addWidget(QLabel("Manual Payload:"))
        self.manual_val = QLineEdit()
        right_vbox.addWidget(self.manual_val)
        self.btn_single = QPushButton("Send Single (Repeater)")
        self.btn_single.clicked.connect(lambda: self.start_attack("single"))
        right_vbox.addWidget(self.btn_single)

        right_vbox.addWidget(QLabel("<hr>Wordlist:"))
        self.wordlist = QListWidget()
        right_vbox.addWidget(self.wordlist)
        
        wl_btns = QHBoxLayout()
        btn_add = QPushButton("Add"); btn_add.clicked.connect(self.add_to_wordlist)
        btn_clr = QPushButton("Clear"); btn_clr.clicked.connect(self.wordlist.clear)
        wl_btns.addWidget(btn_add); wl_btns.addWidget(btn_clr)
        right_vbox.addLayout(wl_btns)

        self.btn_brute = QPushButton("RUN WORDLIST")
        self.btn_brute.clicked.connect(lambda: self.start_attack("brute"))
        right_vbox.addWidget(self.btn_brute)
        
        right_vbox.addStretch()
        main_layout.addWidget(right_panel, 1)

    def mark_payload_selection(self):
        cursor = self.req_edit.textCursor()
        if cursor.hasSelection():
            txt = cursor.selectedText()
            cursor.insertText(f"§{txt}§")
        else:
            cursor.insertText("§§")

    def add_to_wordlist(self):
        text, ok = QInputDialog.getText(self, "Payload", "Veri girin:")
        if ok and text: self.wordlist.addItem(text)

    def start_attack(self, mode):
        raw_req = self.req_edit.toPlainText()
        if not raw_req.strip(): return

        payloads = []
        if mode == "single":
            payloads = [self.manual_val.text()]
            self.btn_single.setEnabled(False)
            self.btn_single.setText("Sending...")
        elif mode == "brute":
            payloads = [self.wordlist.item(i).text() for i in range(self.wordlist.count())]

        for p in payloads:
            threading.Thread(target=self.worker_thread, args=(raw_req, p, mode), daemon=True).start()

    def worker_thread(self, raw, p, mode):
        # UI Güncelleme için yardımcı fonksiyon
        use_https = self.check_https.isChecked()
        formatted = RequestParser.parse_to_dict(raw, p, use_https)
        
        if "error" in formatted:
            QTimer.singleShot(0, lambda: self.update_ui(formatted, mode))
        else:
            response = RequestSender.send(formatted)
            QTimer.singleShot(0, lambda: self.update_ui(response, mode))

    def update_ui(self, res, mode):
        if mode == "single":
            self.btn_single.setEnabled(True)
            self.btn_single.setText("Send Single (Repeater)")

        if "error" in res:
            self.res_code.setPlainText(f"HATA: {res['error']}")
            return

        # Sonuçları göster
        self.res_code.setPlainText(res.get("text", ""))
        self.res_web.setHtml(res.get("text", ""))
        headers_txt = "\n".join([f"{k}: {v}" for k, v in res.get("headers", {}).items()])
        self.res_headers.setPlainText(headers_txt)

        row = self.results_table.rowCount()
        self.results_table.insertRow(row)
        self.results_table.setItem(row, 0, QTableWidgetItem(str(row + 1)))
        self.results_table.setItem(row, 1, QTableWidgetItem(str(res.get("payload", ""))))
        self.results_table.setItem(row, 2, QTableWidgetItem(str(res.get("status", "ERR"))))
        self.results_table.setItem(row, 3, QTableWidgetItem(f"{res.get('length', 0)}"))
        self.results_table.setItem(row, 4, QTableWidgetItem(str(res.get("time_ms", 0))))
        self.results_table.scrollToBottom()

if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = PentestTool()
    window.show()
    sys.exit(app.exec())