# Will ToDo

An elegant, minimalist, and strict local-first ToDo application customized for UGNAS App Ecosystem. 

[![License](https://img.shields.io/badge/License-AGPL%203.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-UGNAS%20Pro-green.svg)](#)

---

## ✨ Features

- **Local-First Architecture**: 100% offline database operations. Absolutely no tracking, no cloud servers, and no leaks.
- **Morandi & Tech Dark Visuals**: Exquisite design optimized for both desktop webviews and immersive focus sessions.
- **Flexible Lifecycles**: Intuitive multi-filter task management (All, Today, Mark, Done, Overdue, and Bin) powered by Flatpickr.js.
- **Complete Sovereignty**: Built-in instant data sandbox reset and robust data exporting utilities.

## 📄 Open Source & Compliance Links

To meet the auditing guidelines of private cloud app stores, the legally binding disclosure indices are listed below:

- **Source Code Hub**: [https://github.com/yourusername/willtodo](https://github.com/yourusername/willtodo)
- **End-User License Agreement (EULA)**: [https://github.com/yourusername/willtodo/blob/main/LICENSE](https://github.com/yourusername/willtodo/blob/main/LICENSE)
- **Privacy Protection Guide**: [https://github.com/yourusername/willtodo/blob/main/PRIVACY.md](https://github.com/yourusername/willtodo/blob/main/PRIVACY.md)

## 🛠️ Project Structure

```text
.
├── go.mod
├── go.sum
├── main.go               # Core Backend Engine
├── project.yaml          # UGNAS Application Metadata
├── rootfs_common/
│   └── www/              # Webview Layer
│       ├── css/
│       ├── js/
│       ├── logo/
│       └── index.html    # Application Front-end UI
└── README.md