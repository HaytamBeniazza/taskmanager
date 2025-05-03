import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { Box, Container } from '@mui/material';
import { SnackbarProvider } from 'notistack';
import Navbar from './components/Navbar';
import TaskList from './components/TaskList';
import TaskForm from './components/TaskForm';
import TaskDetails from './components/TaskDetails';
import Dashboard from './components/Dashboard';

const App: React.FC = () => {
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={3000}>
      <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
        <Navbar />
        <Container component="main" sx={{ flex: 1, py: 4 }}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/tasks" element={<TaskList />} />
            <Route path="/tasks/new" element={<TaskForm />} />
            <Route path="/tasks/:id" element={<TaskDetails />} />
            <Route path="/tasks/:id/edit" element={<TaskForm />} />
          </Routes>
        </Container>
      </Box>
    </SnackbarProvider>
  );
};

export default App; 