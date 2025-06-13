import { useDeferredValue, useState } from 'react';

import '@xyflow/react/dist/style.css';

import { Box, Button, Flex, Text } from '@radix-ui/themes';

import type { WorkflowFormData } from './types';

import { WORKFLOW_EDGES, WORKFLOW_NODES } from './constants';

import { useExecuteWorkflow } from './hooks/useExecuteWorkflow';
import { useWorkflow } from './hooks/useWorkflow';

import { ExecutionResultsComponent } from './components/ExecutionResults';
import { UserInputForm } from './components/UserInputForm';
import { WorkflowDiagram } from './components/WorkflowDiagram';

const WORKFLOW_ID = 'weather-alert-workflow';

function App() {
  const {
    nodes,
    edges,
    setNodes,
    setEdges,
    loading: graphLoading,
    error: graphError,
  } = useWorkflow(WORKFLOW_ID);
  const {
    execute,
    results: executionResults,
    loading: isExecuting,
    resetExecuteResult,
  } = useExecuteWorkflow(WORKFLOW_ID);

  // Defer the heavy graph updates
  const deferredNodes = useDeferredValue(nodes);
  const deferredEdges = useDeferredValue(edges);

  // const [isExecuting, setIsExecuting] = useState(false);
  // const [executionResults, setExecutionResults] = useState<Results | null>(null);
  const [formData, setFormData] = useState<WorkflowFormData | null>(null);

  // const executeWorkflow = (data: WorkflowFormData) => {
  //   setFormData(data);
  //   // setIsExecuting(true);

  //   // TODO: Replace with real API call
  //   // const response = await fetch('/api/workflow/execute', {
  //   //   method: 'POST',
  //   //   headers: { 'Content-Type': 'application/json' },
  //   //   body: JSON.stringify({ formData: data })
  //   // });

  //   // Mock execution for demo
  //   setTimeout(() => {
  //     const mockResults: Results = {
  //       executionId: `exec_${Date.now()}`,
  //       status: 'completed',
  //       startTime: new Date().toISOString(),
  //       endTime: new Date().toISOString(),
  //       steps: [
  //         {
  //           stepNumber: 1,
  //           nodeType: 'form',
  //           status: 'success',
  //           duration: 10,
  //           timestamp: new Date().toISOString(),
  //           output: {
  //             message: `Configuration collected for ${data.name}`,
  //             details: data,
  //           },
  //         },
  //         {
  //           stepNumber: 2,
  //           nodeType: 'integration',
  //           status: 'success',
  //           duration: 245,
  //           timestamp: new Date().toISOString(),
  //           output: {
  //             message: `Weather API: ${data.city} = 28°C`,
  //             details: {
  //               city: data.city,
  //               temperature: 28,
  //               endpoint: `https://api.open-meteo.com/v1/forecast?q=${data.city}`,
  //             },
  //           },
  //         },
  //         {
  //           stepNumber: 3,
  //           nodeType: 'condition',
  //           status: 'success',
  //           duration: 5,
  //           timestamp: new Date().toISOString(),
  //           output: {
  //             message: `Condition: 28°C ${data.operator.replace('_', ' ')} ${data.threshold}°C = ${data.operator === 'greater_than' ? 28 > data.threshold : 28 < data.threshold}`,
  //             details: {
  //               temperature: 28,
  //               operator: data.operator,
  //               threshold: data.threshold,
  //               result:
  //                 data.operator === 'greater_than' ? 28 > data.threshold : 28 < data.threshold,
  //             },
  //           },
  //         },
  //         {
  //           stepNumber: 4,
  //           nodeType: 'email',
  //           status: 'success',
  //           duration: 15,
  //           timestamp: new Date().toISOString(),
  //           output: {
  //             message: `Email content prepared for ${data.email}`,
  //             emailContent: {
  //               to: data.email,
  //               subject: `Weather Alert: ${data.city} is 28°C!`,
  //               body: `Hi ${data.name},\n\nWeather alert for ${data.city}!\nCurrent temperature: 28°C\nYour condition: temperature ${data.operator.replace('_', ' ')} ${data.threshold}°C\n\nStay safe!\n\nWeather Alert System`,
  //             },
  //           },
  //         },
  //       ],
  //     };
  //     setExecutionResults(mockResults);
  //     setIsExecuting(false);
  //   }, 2000);
  // };

  const handleExecute = async (data: WorkflowFormData) => {
    setFormData(data);
    await execute(data);
  };

  const onReset = () => {
    resetExecuteResult();
    setFormData(null);
    setNodes(WORKFLOW_NODES);
    setEdges(WORKFLOW_EDGES);
  };

  return (
    <Box style={{ height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <Box
        p="4"
        style={{ borderBottom: '1px solid var(--gray-6)', backgroundColor: 'var(--gray-2)' }}
      >
        <Flex justify="between" align="center">
          <Text size="6" weight="bold" style={{ display: 'block' }}>
            🌤️ Weather Alert Workflow Engine
          </Text>

          <Flex gap="2">
            <Button variant="soft" onClick={onReset}>
              Reset
            </Button>
          </Flex>
        </Flex>
      </Box>

      <Box style={{ flex: 1, display: 'flex', minHeight: 0 }}>
        {/* Left: Workflow Diagram */}
        <Box style={{ flex: 1, minHeight: 0 }} p="4">
          <Text size="4" weight="medium" mb="3">
            Workflow Diagram
          </Text>

          {graphLoading ? (
            <Text>Loading workflow…</Text>
          ) : graphError ? (
            <Text color="red">Error loading workflow: {graphError}</Text>
          ) : (
            <WorkflowDiagram nodes={deferredNodes} edges={deferredEdges} onNodesChange={setNodes} />
          )}
        </Box>

        <Box
          style={{
            borderLeft: '1px solid var(--gray-6)',
            backgroundColor: 'var(--gray-1)',
            width: '400px',
            height: 'calc(100vh - 80px)',
            overflow: 'hidden',
          }}
        >
          {!executionResults ? (
            <UserInputForm onExecute={handleExecute} isExecuting={isExecuting} />
          ) : (
            <ExecutionResultsComponent results={executionResults} formData={formData} />
          )}
        </Box>
      </Box>
    </Box>
  );
}

export default App;
