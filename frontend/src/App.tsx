import { Routes, Route, Navigate } from 'react-router-dom';
import { useLocalStorage } from '@mantine/hooks';
import Layout from './Layout';
import Login from './pages/Login';
import Register from './pages/Register';
import MeetingsList from './pages/MeetingsList';
import MeetingNew from './pages/MeetingNew';
import MeetingPage from './pages/MeetingPage';
import ParticipantPage from './pages/ParticipantPage';

function App() {
  const [token] = useLocalStorage<string | null>({ key: 'token', defaultValue: null });

  return (
    <Routes>
      <Route path="/" element={<Layout token={token} />}>
        <Route index element={<Navigate to={token ? '/meetings' : '/login'} replace />} />
        <Route path="login" element={token ? <Navigate to="/meetings" replace /> : <Login />} />
        <Route path="register" element={token ? <Navigate to="/meetings" replace /> : <Register />} />
        <Route path="meetings" element={token ? <MeetingsList /> : <Navigate to="/login" replace />} />
        <Route path="meetings/new" element={token ? <MeetingNew /> : <Navigate to="/login" replace />} />
        <Route path="meetings/:id" element={<MeetingPage token={token} />} />
        <Route path="meetings/:id/participate" element={<ParticipantPage />} />
        <Route path="meetings/:id/participate/:participantId" element={<ParticipantPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default App;
