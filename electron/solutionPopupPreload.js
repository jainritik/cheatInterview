const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('solutionPopupAPI', {
  onSolutionData: (callback) => ipcRenderer.on('solution-data', (event, ...args) => callback(...args)),
});
console.log('Solution Popup Preload script loaded');
