#!/bin/bash
# a self-extracting script header
# Credit: https://www.linuxjournal.com/node/1005818
set -e
echo "FastForward Server installer v0.1.0"
if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

mkdir -p ./installer_temp

# determine the line number of this script where the zip begins
ZIP_LINE=`awk '/^__ZIP_BELOW__/ {print NR + 1; exit 0; }' $0`

# use the tail command and the line number we just determined to skip
# past this leading script code and pipe the zipfile to tar
echo "Extracting..."
tail -n+$ZIP_LINE $0 | tar xz -C ./installer_temp

# Run following after extracting
cd ./installer_temp
usr=$SUDO_USER
USER_HOME=$(getent passwd $usr | cut -d: -f6)
ffdir=$USER_HOME/.fastforward
echo "Moving executables..."
cp fastforward /usr/local/bin/

echo "Moving config files..."
mkdir -p /etc/fastforward
cp cfg/.env /etc/fastforward/
sed -i -e "s|fullpathofhomehere|$USER_HOME|g" /etc/fastforward/.env
cp cfg/ip_list.txt $ffdir/

echo "Creating systemd service..."
mkdir -p /etc/systemd/system
cp dep/fastforward.service /etc/systemd/system/
systemctl enable fastforward

echo "Cleaning up..."
cd ..
sudo chown -R $USR: $USER_HOME/.fastforward
rm -rf ./installer_temp

while true; do
    read -p "Do you wish to configure the server now?" yn
    case $yn in
        [Yy]* ) nano /etc/fastforward/.env; break;;
        [Nn]* ) exit 0;;
        * ) echo "Please answer yes or no.";;
    esac
done

echo "All done, use 'systemctl start fastforward' to start server now. MAKE SURE YOU HAVE YOUR DATABASE SETUP"
exit 0
# the 'exit 0' immediately above prevents this line from being executed

__ZIP_BELOW__
