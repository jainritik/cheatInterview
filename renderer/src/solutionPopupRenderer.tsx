import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
// Assuming Prism and dracula style are used, like in Solutions.tsx
// You might need to install these if not already project-wide dependencies for renderer
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { dracula } from 'react-syntax-highlighter/dist/esm/styles/prism';

interface SolutionData {
  code: string;
  explanation: string;
  // Add other fields if they exist in the solutionData object
}

// Augment the Window interface to include solutionPopupAPI
declare global {
  interface Window {
    solutionPopupAPI?: {
      onSolutionData: (callback: (data: any) => void) => void;
    };
  }
}

const SolutionDisplay: React.FC = () => {
  const [solution, setSolution] = useState<SolutionData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (window.solutionPopupAPI && window.solutionPopupAPI.onSolutionData) {
      window.solutionPopupAPI.onSolutionData((data: any) => {
        console.log('Solution data received in popup:', data);
        // The data might be directly the solution or nested under a 'data' property
        const actualSolutionData = data?.data && typeof data.data === 'object' ? data.data : data;

        if (actualSolutionData && actualSolutionData.code) {
          setSolution(actualSolutionData as SolutionData);
        } else {
          console.error('Received malformed or incomplete solution data:', actualSolutionData);
          setError('Received malformed or incomplete solution data. Ensure `code` property exists and is not null.');
        }
      });
    } else {
      console.error('solutionPopupAPI not found. Was the preload script loaded?');
      setError('Error: Preload script not loaded. Cannot receive solution data.');
    }
  }, []);

  if (error) {
    return <div style={{ padding: '20px', color: 'red', backgroundColor: '#ffe0e0', border: '1px solid red', borderRadius: '4px' }}>{error}</div>;
  }

  if (!solution) {
    return <div style={{ padding: '20px', textAlign: 'center' }}>Loading solution...</div>;
  }

  return (
    <div className="solution-container">
      {solution.explanation && (
        <div className="explanation-section">
          <h2>Explanation</h2>
          <p>{solution.explanation}</p>
        </div>
      )}
      {solution.code && (
        <div className="code-section">
          <h2>Code</h2>
          <SyntaxHighlighter language="cpp" style={dracula} showLineNumbers customStyle={{ margin: 0 }}>
            {solution.code}
          </SyntaxHighlighter>
        </div>
      )}
      {!solution.code && !solution.explanation && (
        <div style={{ padding: '20px', textAlign: 'center' }}>No content to display in the solution.</div>
      )}
    </div>
  );
};

const rootElement = document.getElementById('root');
if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <React.StrictMode>
      <SolutionDisplay />
    </React.StrictMode>
  );
} else {
  console.error('Root element not found in solutionPopup.html. Solution UI cannot be mounted.');
  const errorDiv = document.createElement('div');
  errorDiv.innerHTML = '<h1 style="color: red; text-align: center;">Critical Error: Root HTML element not found.</h1>';
  document.body.appendChild(errorDiv);
}
