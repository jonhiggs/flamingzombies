install:
	cp -f  bin/* /usr/local/bin/
	cp -rf libexec/flamingzombies /usr/local/libexec
	cp -f  man/man1/* /usr/local/man/man1/
	cp -f  man/man5/* /usr/local/man/man5/
	cp -f  man/man7/* /usr/local/man/man7/
	cp -f  scripts/openbsd_rc /etc/rc.d/fz
	chmod 775 /etc/rc.d/fz
	chown root:wheel /etc/rc.d/fz
