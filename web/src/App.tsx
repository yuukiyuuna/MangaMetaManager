import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import LibraryPage from './pages/LibraryPage';
import SettingsPage from './pages/SettingsPage';
import { ToastContainer } from './components/Toast';
import TaskStatus from './components/TaskStatus';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<LibraryPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Routes>
      </Layout>
      <ToastContainer />
      <TaskStatus />
    </Router>
  );
}

export default App;
