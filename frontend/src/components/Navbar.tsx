import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  IconButton,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  List as ListIcon,
  Add as AddIcon,
} from '@mui/icons-material';

const Navbar: React.FC = () => {
  return (
    <AppBar position="static">
      <Toolbar>
        <Typography
          variant="h6"
          component={RouterLink}
          to="/"
          sx={{
            flexGrow: 1,
            textDecoration: 'none',
            color: 'inherit',
            display: 'flex',
            alignItems: 'center',
          }}
        >
          Task Manager
        </Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            color="inherit"
            component={RouterLink}
            to="/"
            startIcon={<DashboardIcon />}
          >
            Dashboard
          </Button>
          <Button
            color="inherit"
            component={RouterLink}
            to="/tasks"
            startIcon={<ListIcon />}
          >
            Tasks
          </Button>
          <Button
            color="inherit"
            component={RouterLink}
            to="/tasks/new"
            startIcon={<AddIcon />}
          >
            New Task
          </Button>
        </Box>
      </Toolbar>
    </AppBar>
  );
};

export default Navbar; 