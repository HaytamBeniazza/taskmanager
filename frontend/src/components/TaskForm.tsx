import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Grid,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  Autocomplete,
  TextField as MuiTextField,
  Stack,
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
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

const TaskForm: React.FC = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [task, setTask] = useState<Partial<Task>>({
    title: '',
    description: '',
    completed: false,
    category: '',
    priority: 'medium',
    tags: [],
  });
  const [dueDate, setDueDate] = useState<Date | null>(null);
  const [availableTags, setAvailableTags] = useState<string[]>([]);

  useEffect(() => {
    fetchAvailableTags();
    if (id) {
      fetchTask();
    }
  }, [id]);

  const fetchTask = async () => {
    try {
      const response = await axios.get(`http://localhost:8080/tasks/${id}`);
      setTask(response.data);
      if (response.data.dueDate) {
        setDueDate(new Date(response.data.dueDate));
      }
    } catch (error) {
      console.error('Error fetching task:', error);
    }
  };

  const fetchAvailableTags = async () => {
    try {
      const response = await axios.get('http://localhost:8080/tasks/tags');
      setAvailableTags(response.data.tags);
    } catch (error) {
      console.error('Error fetching tags:', error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const taskData = {
        ...task,
        dueDate: dueDate?.toISOString(),
      };

      if (id) {
        await axios.put(`http://localhost:8080/tasks/${id}`, taskData);
      } else {
        await axios.post('http://localhost:8080/tasks', taskData);
      }
      navigate('/tasks');
    } catch (error) {
      console.error('Error saving task:', error);
    }
  };

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 4 }}>
        {id ? 'Edit Task' : 'New Task'}
      </Typography>

      <Card>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <Grid container spacing={3}>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Title"
                  value={task.title}
                  onChange={(e) => setTask({ ...task, title: e.target.value })}
                  required
                />
              </Grid>

              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Description"
                  value={task.description}
                  onChange={(e) => setTask({ ...task, description: e.target.value })}
                  multiline
                  rows={4}
                />
              </Grid>

              <Grid item xs={12} sm={6}>
                <FormControl fullWidth>
                  <InputLabel>Category</InputLabel>
                  <Select
                    value={task.category}
                    onChange={(e) => setTask({ ...task, category: e.target.value })}
                    label="Category"
                  >
                    <MenuItem value="Work">Work</MenuItem>
                    <MenuItem value="Personal">Personal</MenuItem>
                    <MenuItem value="Learning">Learning</MenuItem>
                  </Select>
                </FormControl>
              </Grid>

              <Grid item xs={12} sm={6}>
                <FormControl fullWidth>
                  <InputLabel>Priority</InputLabel>
                  <Select
                    value={task.priority}
                    onChange={(e) => setTask({ ...task, priority: e.target.value as 'low' | 'medium' | 'high' })}
                    label="Priority"
                  >
                    <MenuItem value="low">Low</MenuItem>
                    <MenuItem value="medium">Medium</MenuItem>
                    <MenuItem value="high">High</MenuItem>
                  </Select>
                </FormControl>
              </Grid>

              <Grid item xs={12}>
                <LocalizationProvider dateAdapter={AdapterDateFns}>
                  <DatePicker
                    label="Due Date"
                    value={dueDate}
                    onChange={(newValue) => setDueDate(newValue)}
                    slotProps={{ textField: { fullWidth: true } }}
                  />
                </LocalizationProvider>
              </Grid>

              <Grid item xs={12}>
                <Autocomplete
                  multiple
                  options={availableTags}
                  value={task.tags || []}
                  onChange={(_, newValue) => setTask({ ...task, tags: newValue })}
                  renderTags={(value, getTagProps) =>
                    value.map((option, index) => (
                      <Chip
                        label={option}
                        {...getTagProps({ index })}
                        key={option}
                      />
                    ))
                  }
                  renderInput={(params) => (
                    <MuiTextField
                      {...params}
                      label="Tags"
                      placeholder="Add tags"
                    />
                  )}
                />
              </Grid>

              <Grid item xs={12}>
                <Stack direction="row" spacing={2} justifyContent="flex-end">
                  <Button
                    variant="outlined"
                    onClick={() => navigate('/tasks')}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    variant="contained"
                    color="primary"
                  >
                    {id ? 'Update' : 'Create'} Task
                  </Button>
                </Stack>
              </Grid>
            </Grid>
          </form>
        </CardContent>
      </Card>
    </Box>
  );
};

export default TaskForm; 