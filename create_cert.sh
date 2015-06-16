#/!bin/sh
TOPDIR=$(pwd)
SSL_KEY_NAME=$(hostname --fqdn)
CONF_FILE=$(mktemp)
OUTPATH=${TOPDIR}/ssl

mkdir -p ${OUTPATH}/private 2>/dev/null
mkdir -p ${OUTPATH}/certificates 2>/dev/null

sed \
  -e "s/@HostName@/${SSL_KEY_NAME}/" \
  -e "s|privkey.pem|/etc/ssl/private/${SSL_KEY_NAME}.pem|" \
      "/usr/share/ssl-cert/ssleay.cnf" > "${CONF_FILE}"
openssl req -config "${CONF_FILE}" -new -x509 -days 3650 \
  -nodes -out "${OUTPATH}/certificates/${SSL_KEY_NAME}.crt" -keyout "${OUTPATH}/private/${SSL_KEY_NAME}.key"

rm "${CONF_FILE}"
chown $(whoami) "${OUTPATH}/private/${SSL_KEY_NAME}.key"
chmod 440 "${OUTPATH}/private/${SSL_KEY_NAME}.key"
