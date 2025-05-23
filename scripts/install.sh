# Set permissions
chown -R asterisk:asterisk /opt/asterisk-call-routing
chown -R asterisk:asterisk /var/log/asterisk-router

# Setup database
echo "Setting up database..."
mysql -u root -ptemppass < migrations/001_initial_schema.sql

# Enable and start service
systemctl daemon-reload
systemctl enable asterisk-router
systemctl start asterisk-router

echo "Installation complete!"
echo "Service status:"
systemctl status asterisk-router
# Set permissions
chown -R asterisk:asterisk /opt/asterisk-call-routing
chown -R asterisk:asterisk /var/log/asterisk-router

# Setup database
echo "Setting up database..."
mysql -u root -ptemppass < migrations/001_initial_schema.sql

# Enable and start service
systemctl daemon-reload
systemctl enable asterisk-router
systemctl start asterisk-router

echo "Installation complete!"
echo "Service status:"
systemctl status asterisk-router
