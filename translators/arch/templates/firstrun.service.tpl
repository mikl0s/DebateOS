[Unit]
Description=DebateOS first-run: {description}
ConditionPathExists=!{flag_file}
After=graphical-session.target

[Service]
Type=oneshot
ExecStart={exec_path}
ExecStartPost=/bin/touch {flag_file}
RemainAfterExit=yes

[Install]
WantedBy=graphical-session.target
