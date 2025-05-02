import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Chip,
  Stack,
  Divider,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import {
  Edit as EditIcon,
  Delete as DeleteIcon,
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
} from '@mui/icons-material';
import axios from 'axios';

interface Task {
  id: number;
  title: string;
  description: string;
  completed: boolean;
  createdAt: string;
  dueDate: string;
  category: string;
  priority: 'low' | 'medium' | 'high';
  tags: string[];
}

const TaskDetails: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [task, setTask] = useState<Task | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  useEffect(() => {
    fetchTask();
  }, [id]);

  const fetchTask = async () => {
    try {
      const response = await axios.get(`http://localhost:8080/tasks/${id}`);
      setTask(response.data);
    } catch (error) {
      console.error('Error fetching task:', error);
    }
  };

  const handleToggleComplete = async () => {
    if (!task) return;
    try {
      await axios.patch(`http://localhost:8080/tasks/${id}`, {
        completed: !task.completed,
      });
      fetchTask();
    } catch (error) {
      console.error('Error updating task:', error);
    }
  };

  const handleDelete = async () => {
    try {
      await axios.delete(`http://localhost:8080/tasks/${id}`);
      navigate('/tasks');
    } catch (error) {
      console.error('Error deleting task:', error);
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high':
        return 'error';
      case 'medium':
        return 'warning';
      case 'low':
        return 'success';
      default:
        return 'default';
    }
  };

  if (!task) {
    return <Typography>Loading...</Typography>;
  }

  return (
    <Box>
      <Stack direction="row" spacing={2} sx={{ mb: 4 }}>
        <Typography variant="h4" component="h1">
          Task Details
        </Typography>
        <Box sx={{ flexGrow: 1 }} />
        <IconButton
          color="primary"
          onClick={() => navigate(`/tasks/${id}/edit`)}
        >
          <EditIcon />
        </IconButton>
        <IconButton
          color="error"
          onClick={() => setDeleteDialogOpen(true)}
        >
          <DeleteIcon />
        </IconButton>
      </Stack>

      <Card>
        <CardContent>
          <Stack spacing={3}>
            <Stack direction="row" spacing={2} alignItems="center">
              <Typography variant="h5" component="h2">
                {task.title}
              </Typography>
              <Chip
                label={task.priority}
                color={getPriorityColor(task.priority)}
                size="small"
              />
              <Chip
                label={task.category}
                variant="outlined"
                size="small"
              />
              <Box sx={{ flexGrow: 1 }} />
              <Button
                variant="outlined"
                startIcon={task.completed ? <CancelIcon /> : <CheckCircleIcon />}
                onClick={handleToggleComplete}
                color={task.completed ? 'error' : 'success'}
              >
                {task.completed ? 'Mark Incomplete' : 'Mark Complete'}
              </Button>
            </Stack>

            <Divider />

            <Typography variant="body1" color="text.secondary">
              {task.description}
            </Typography>

            <Stack direction="row" spacing={1}>
              {task.tags.map((tag) => (
                <Chip
                  key={tag}
                  label={tag}
                  size="small"
                  variant="outlined"
                />
              ))}
            </Stack>

            <Stack direction="row" spacing={2} color="text.secondary">
              <Typography variant="body2">
                Created: {new Date(task.createdAt).toLocaleDateString()}
              </Typography>
              {task.dueDate && (
                <Typography variant="body2">
                  Due: {new Date(task.dueDate).toLocaleDateString()}
                </Typography>
              )}
            </Stack>
          </Stack>
        </CardContent>
      </Card>

      <Dialog
        open={deleteDialogOpen}
        onClose={() => setDeleteDialogOpen(false)}
      >
        <DialogTitle>Delete Task</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete this task? This action cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} color="error">
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default TaskDetails; 