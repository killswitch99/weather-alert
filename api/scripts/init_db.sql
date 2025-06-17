-- Ensure correct schema is set
SET search_path TO public;
-- Insert sample workflow if it doesn't exist
INSERT INTO workflows (id, name)
SELECT 
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'Weather Alert Workflow'
WHERE NOT EXISTS (
    SELECT 1 FROM workflows WHERE id = '550e8400-e29b-41d4-a716-446655440000'::uuid
);

-- Insert nodes and store their IDs
WITH node_inserts AS (
    INSERT INTO workflow_nodes (workflow_id, node_id, node_type, position_x, position_y, label, description, metadata)
    SELECT * FROM (
        SELECT 
            '550e8400-e29b-41d4-a716-446655440000'::uuid,
            'start',
            'start',
            -160,
            300,
            'Start',
            'Begin weather check workflow',
            '{"hasHandles": {"source": true, "target": false}}'::jsonb
        WHERE NOT EXISTS (
            SELECT 1 FROM workflow_nodes 
            WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
            AND node_type = 'start'
        )
        UNION ALL
        SELECT 
            '550e8400-e29b-41d4-a716-446655440000'::uuid,
            'form',
            'form',
            152,
            304,
            'User Input',
            'Process collected data - name, email, location',
            '{
                "hasHandles": {"source": true, "target": true},
                "inputFields": ["name", "email", "city"],
                "outputVariables": ["name", "email", "city"]
            }'::jsonb
        WHERE NOT EXISTS (
            SELECT 1 FROM workflow_nodes 
            WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
            AND node_type = 'form'
        )
        UNION ALL
        SELECT 
            '550e8400-e29b-41d4-a716-446655440000'::uuid,
            'weather-api',
            'integration',
            460,
            304,
            'Weather API',
            'Fetch current temperature for {{city}}',
            '{
                "hasHandles": {"source": true, "target": true},
                "inputVariables": ["city"],
                "apiEndpoint": "https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true",
                "options": [
                    {"city": "Sydney", "lat": -33.8688, "lon": 151.2093},
                    {"city": "Melbourne", "lat": -37.8136, "lon": 144.9631},
                    {"city": "Brisbane", "lat": -27.4698, "lon": 153.0251},
                    {"city": "Perth", "lat": -31.9505, "lon": 115.8605},
                    {"city": "Adelaide", "lat": -34.9285, "lon": 138.6007}
                ],
                "outputVariables": ["temperature"]
            }'::jsonb
        WHERE NOT EXISTS (
            SELECT 1 FROM workflow_nodes 
            WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
            AND node_type = 'integration'
        )
        UNION ALL
        SELECT 
            '550e8400-e29b-41d4-a716-446655440000'::uuid,
            'condition',
            'condition',
            794,
            304,
            'Check Condition',
            'Evaluate temperature threshold',
            '{
                "hasHandles": {"source": ["true", "false"], "target": true},
                "conditionExpression": "temperature {{operator}} {{threshold}}",
                "outputVariables": ["conditionMet"]
            }'::jsonb
        WHERE NOT EXISTS (
            SELECT 1 FROM workflow_nodes 
            WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
            AND node_type = 'condition'
        )
        UNION ALL
        SELECT 
            '550e8400-e29b-41d4-a716-446655440000'::uuid,
            'email',
            'email',
            1096,
            88,
            'Send Alert',
            'Email weather alert notification',
            '{
                "hasHandles": {"source": true, "target": true},
                "inputVariables": ["name", "city", "temperature"],
                "emailTemplate": {
                    "subject": "Weather Alert",
                    "body": "Weather alert for {{city}}! Temperature is {{temperature}}°C!"
                },
                "outputVariables": ["emailSent"]
            }'::jsonb
        WHERE NOT EXISTS (
            SELECT 1 FROM workflow_nodes 
            WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
            AND node_type = 'email'
        )
        UNION ALL
        SELECT 
            '550e8400-e29b-41d4-a716-446655440000'::uuid,
            'end',
            'end',
            1360,
            302,
            'Complete',
            'Workflow execution finished',
            '{"hasHandles": {"source": false, "target": true}}'::jsonb
        WHERE NOT EXISTS (
            SELECT 1 FROM workflow_nodes 
            WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
            AND node_type = 'end'
        )
    ) AS nodes
    RETURNING id, node_type
)
-- Insert edges using node types
INSERT INTO workflow_edges (
    workflow_id, 
    source_node_id, 
    target_node_id,
    edge_id, 
    type, 
    animated, 
    stroke_color, 
    stroke_width, 
    label, 
    source_handle, 
    label_style
)
-- Edge 1: Start to Form
SELECT 
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'start',
    'form',
    'e1',
    'smoothstep', 
    true,
    '#10b981',
    3,
    'Initialize',
    '',
    '{}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM workflow_edges 
    WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
    AND edge_id = 'e1'
)
UNION ALL
-- Edge 2: Form to Weather API
SELECT 
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'form',
    'weather-api',
    'e2',
    'smoothstep', 
    true,
    '#3b82f6',
    3,
    'Submit Data',
    '',
    '{}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM workflow_edges 
    WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
    AND edge_id = 'e2'
)
UNION ALL
-- Edge 3: Weather API to Condition
SELECT 
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'weather-api',
    'condition',
    'e3',
    'smoothstep', 
    true,
    '#f97316',
    3,
    'Temperature Data',
    '',
    '{}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM workflow_edges 
    WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
    AND edge_id = 'e3'
)
UNION ALL
-- Edge 4: Condition to Email (true path)
SELECT 
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'condition',
    'email',
    'e4',
    'smoothstep', 
    true,
    '#10b981',
    3,
    '✓ Condition Met',
    'true',
    '{"fill": "#10b981", "fontWeight": "bold"}'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM workflow_edges 
    WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
    AND edge_id = 'e4'
)
UNION ALL
-- Edge 5: Condition to End (false path)
SELECT 
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'condition',
    'end',
    'e5',
    'smoothstep', 
    true,
    '#6b7280',
    3,
    '✗ No Alert Needed',
    'false',
    '{"fill": "#6b7280", "fontWeight": "bold"}'::jsonb
FROM node_inserts source
JOIN node_inserts target ON target.node_type = 'end'
WHERE source.node_type = 'condition'
AND NOT EXISTS (
    SELECT 1 FROM workflow_edges 
    WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
    AND edge_id = 'e5'
)
UNION ALL
SELECT 
    '550e8400-e29b-41d4-a716-446655440000'::uuid,
    'email',
    'end',
    'e6',
    'smoothstep', 
    true,
    '#ef4444',
    2,
    'Alert Sent',
    '',
    '{"fill": "#ef4444", "fontWeight": "bold"}'::jsonb
FROM node_inserts source
JOIN node_inserts target ON target.node_type = 'end'
WHERE source.node_type = 'email'
AND NOT EXISTS (
    SELECT 1 FROM workflow_edges 
    WHERE workflow_id = '550e8400-e29b-41d4-a716-446655440000'::uuid 
    AND edge_id = 'e6'
); 