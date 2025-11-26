CREATE TABLE IF NOT EXISTS external_alerts (
                                               id SERIAL PRIMARY KEY,
                                               source VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_source CHECK (source IN ('siem-1', 'siem-2', 'firewall', 'ids', 'antivirus', 'endpoint', 'cloud-security', 'email-gateway', 'network-monitor', 'vulnerability-scanner'))
    );

-- Insert sample data
INSERT INTO external_alerts (source, severity, description, created_at) VALUES
                                                                            ('siem-1', 'high', 'Suspicious login attempt detected', NOW() - INTERVAL '1 hour'),
                                                                            ('firewall', 'critical', 'Multiple blocked intrusion attempts', NOW() - INTERVAL '2 hours'),
                                                                            ('ids', 'medium', 'Unusual network traffic pattern', NOW() - INTERVAL '3 hours'),
                                                                            ('antivirus', 'low', 'Potentially unwanted program detected', NOW() - INTERVAL '4 hours'),
                                                                            ('endpoint', 'high', 'Unauthorized software installation', NOW() - INTERVAL '5 hours'),
                                                                            ('cloud-security', 'critical', 'Exposed S3 bucket detected', NOW() - INTERVAL '6 hours'),
                                                                            ('email-gateway', 'medium', 'Phishing email blocked', NOW() - INTERVAL '7 hours'),
                                                                            ('network-monitor', 'low', 'High bandwidth usage detected', NOW() - INTERVAL '8 hours'),
                                                                            ('vulnerability-scanner', 'high', 'Critical CVE found in production', NOW() - INTERVAL '9 hours'),
                                                                            ('siem-2', 'medium', 'Failed authentication attempts', NOW() - INTERVAL '10 hours');