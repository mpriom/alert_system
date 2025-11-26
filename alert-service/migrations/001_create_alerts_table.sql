-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source VARCHAR(255) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    whole_event BYTEA NOT NULL,
    enrichment_type VARCHAR(100),
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

-- Create index on severity for faster queries
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);

-- Create index on created_at for faster sorting
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);

-- Create index on source
CREATE INDEX IF NOT EXISTS idx_alerts_source ON alerts(source);

-- Insert some sample data with varying dates and whole_event blobs
INSERT INTO alerts (source, severity, description, whole_event, created_at) VALUES
                                                                                (
                                                                                    'siem-1',
                                                                                    'high',
                                                                                    'Suspicious login detected from unknown IP',
                                                                                    E'\\x7b2274686972645f70617274795f7575696422203a2022333830306664323930612d356165372d346265622d383730322d333661313936303934396530222c20226576656e745f74797065223a20226c6f67696e5f617474656d7074222c2022736f75726365223a2022656d61696c222c2022697022203a2022203139322e3136382e312e323535222c2022757365725f6167656e74223a20224d6f7a696c6c612f352e302028583131293b204c696e75782078383620222c20227374617475733a2022737573706963696f7573222c20227269736b5f73636f7265223a20227b2273636f7265223a203835207d227d'::bytea,
                                                                                    NOW() - INTERVAL '1 hour'
                                                                                ),
                                                                                (
                                                                                    'siem-1',
                                                                                    'critical',
                                                                                    'Multiple failed authentication attempts',
                                                                                    E'\\x7b2274686972645f70617274795f7575696422203a20223638626363386137382d383135382d343638662d623964392d636134633663316439633738222c20226576656e745f74797065223a20226661696c65645f617574686e222c2022736f75726365223a20227373682d736572766572222c2022697022203a2022313032382e3230302e302e3135222c2022757365726e616d6522203a202022726f6f74222c2022617474656d7074735f636f756e74223a20352c202272697461656b5f73636f7265223a20227b2273636f7265223a203935207d227d'::bytea,
                                                                                    NOW() - INTERVAL '2 days'
                                                                                ),
                                                                                (
                                                                                    'firewall-1',
                                                                                    'medium',
                                                                                    'Unusual outbound traffic pattern detected',
                                                                                    E'\\x7b2274686972645f70617274795f7575696422203a20223433396432303335322d346239372d346536662d616331302d366633656233303231346230222c20226576656e745f74797065223a20226e6574776f726b5f7472616666696322222c2022736f75726365223a2022666972657761696c222c2022646573745f6970223a2022203130382e3138362e3230302e3235222c2022627974655f636f756e74223a2022206d2037333938202037383739222c20227269736b5f73636f7265223a20227b2273636f7265223a203630207d227d'::bytea,
                                                                                    NOW() - INTERVAL '5 days'
                                                                                ),
                                                                                (
                                                                                    'ids-1',
                                                                                    'high',
                                                                                    'Potential SQL injection attempt blocked',
                                                                                    E'\\x7b2274686972645f70617274795f7575696422203a20226561656134636362622d646135612d343232642d613434362d653731373563383233356638222c20226576656e745f74797065223a20227765625f61747461636b222c2022736f75726365223a2022776166222c2022726571756573745f75726922203a2022203132332f757365725f6c6f67696e2e7068703f75736572696428273b27223a20222044524f50207461626c652075736572733b272d2d222c20226174746163685f74797065223a202273716c5f696e6a656374696f6e222c20226374696f6e223a2022626c6f636b2e207d227d'::bytea,
                                                                                    NOW() - INTERVAL '7 days'
                                                                                ),
                                                                                (
                                                                                    'siem-2',
                                                                                    'low',
                                                                                    'User account locked due to failed logins',
                                                                                    E'\\x7b2274686972645f70617274795f7575696422203a20226637346162656265302d303939612d343635312d613335332d643937373332396633626130222c20226576656e745f74797065223a20226163636f756e745f6c6f636b222c2022736f75726365223a20226164222c20222075736572223a20226a6f686e2e646f65222c20226661696c65645f617474656d707473223a20332c20226c6f636b5f64757261746964223a2022333020696e75746573222c2022706b5f73636f7265223a20227b2273636f7265223a203230207d227d'::bytea,
                                                                                    NOW() - INTERVAL '10 days'
                                                                                );