#!/bin/sh

# We need to bind mount in /etc/ from the host (Linux agent expects this to be
# accessible for a variety of reasons, such as user management and SSH key
# management), so we need to ensure these files come in after the fact and
# don't get over-written.
cp /opt/waagent.conf /etc/waagent.conf
cp /opt/sshd_config /etc/ssh/sshd_config
cp /opt/openssl.cnf /etc/ssl/openssl.cnf
cp /opt/motd /etc/motd
cp /opt/waagent.logrotate /etc/logrotate.d/waagent.logrotate

# TODO: Maybe this would be better done somewhere else, but this will get the
# job done for now.
chown -R docker:docker /var/log

# Azure wants this setting.
echo "ClientAliveInterval 180" >>/etc/ssh/sshd_config

# Workaround due to 'gawk' name.
# Azure agent just hardcodes shelling out to 'awk' and it chokes on Busybox awk
# for some things.
rm /usr/bin/awk
ln -s /usr/bin/gawk /usr/bin/awk

echo '
CHFN_RESTRICT    rwh
DEFAULT_HOME     yes
ENCRYPT_METHOD   SHA512
ENV_PATH         PATH=/usr/local/bin:/usr/bin:/bin:/usr/local/games:/usr/games
ENV_SUPATH       PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ERASECHAR        0177
FAILLOG_ENAB     yes
FTMP_FILE        /var/log/btmp
GID_MAX          60000
GID_MIN          1000
HUSHLOGIN_FILE   .hushlogin
KILLCHAR         025
LOGIN_RETRIES    5
LOGIN_TIMEOUT    60
LOG_OK_LOGINS    no
LOG_UNKFAIL_ENAB no
MAIL_DIR         /var/mail
PASS_MAX_DAYS    99999
PASS_MIN_DAYS    0
PASS_WARN_AGE    7
SU_NAME          su
SYSLOG_SG_ENAB   yes
SYSLOG_SU_ENAB   yes
TTYGROUP         tty
TTYPERM          0600
UID_MAX          60000
UID_MIN          1000
UMASK            022
USERGROUPS_ENAB yes' >/etc/login.defs

/usr/sbin/waagent -daemon -verbose

# Updated apk packages and add sudo to sync with host
apk --update add sudo