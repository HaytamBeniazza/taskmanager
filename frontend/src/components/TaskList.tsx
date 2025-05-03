import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Grid,
  Chip,
  TextField,
  InputAdornment,
  IconButton,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Checkbox,
  TableContainer,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  Paper,
  Divider,
  FormControlLabel,
  Switch,
  Stack,
  Pagination,
  Tooltip,
  Avatar,
  Badge,
  Alert,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Tabs,
  Tab,
  Snackbar,
} from '@mui/material';
import {
  Search as SearchIcon,
  Add as AddIcon,
  Sort as SortIcon,
  Delete as DeleteIcon,
  Edit as EditIcon,
  FilterList as FilterListIcon,
  Check as CheckIcon,
  MoreVert as MoreVertIcon,
  ViewList as ViewListIcon,
  ViewModule as ViewModuleIcon,
  Refresh as RefreshIcon,
  Tune as TuneIcon,
  Done as DoneIcon,
  Flag as FlagIcon,
  Today as TodayIcon,
  Error as ErrorIcon,
  CheckCircle as CheckCircleIcon,
  Warning as WarningIcon,
  DateRange as DateRangeIcon,
} from '@mui/icons-material';
import axios from 'axios';
import { format, isAfter, isBefore, parseISO } from 'date-fns';

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
  notes: any[];
  subtasks: any[];
  reminders: any[];
}

interface PaginationInfo {
  total: number;
  currentPage: number;
  perPage: number;
  totalPages: number;
  hasMore: boolean;
}

const TaskList: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [selectedTasks, setSelectedTasks] = useState<number[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [category, setCategory] = useState('');
  const [categories, setCategories] = useState<string[]>([]);
  const [priority, setPriority] = useState('');
  const [tag, setTag] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [completed, setCompleted] = useState('');
  const [sortBy, setSortBy] = useState('dueDate');
  const [sortOrder, setSortOrder] = useState('asc');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [page, setPage] = useState(1);
  const [perPage, setPerPage] = useState(9);
  const [paginationInfo, setPaginationInfo] = useState<PaginationInfo | null>(null);
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const [confirmCompleteOpen, setConfirmCompleteOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [snackbar, setSnackbar] = useState<{open: boolean, message: string, severity: 'success' | 'error'}>({
    open: false,
    message: '',
    severity: 'success'
  });
  const [dueBefore, setDueBefore] = useState('');
  const [dueAfter, setDueAfter] = useState('');
  const [showUpcoming, setShowUpcoming] = useState(false);
  const [showOverdue, setShowOverdue] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetchCategories();
    fetchTags();
    fetchTasks();
  }, [sortBy, sortOrder, page, perPage, showUpcoming, showOverdue]);

  const fetchCategories = async () => {
    try {
      const response = await axios.get('http://localhost:8080/tasks/categories');
      setCategories(response.data.categories || []);
    } catch (error) {
      console.error('Error fetching categories:', error);
    }
  };

  const fetchTags = async () => {
    try {
      const response = await axios.get('http://localhost:8080/tasks/tags');
      setTags(response.data.tags || []);
    } catch (error) {
      console.error('Error fetching tags:', error);
    }
  };

  const fetchTasks = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Build the query parameters
      const params = new URLSearchParams();
      params.append('sortBy', sortBy);
      params.append('sortDir', sortOrder);
      params.append('page', page.toString());
      params.append('perPage', perPage.toString());
      
      if (showUpcoming) {
        // Current date in ISO format
        const now = new Date().toISOString();
        params.append('dueAfter', now);
        params.append('completed', 'false');
      }
      
      if (showOverdue) {
        // Current date in ISO format
        const now = new Date().toISOString();
        params.append('dueBefore', now);
        params.append('completed', 'false');
      }

      const response = await axios.get(`http://localhost:8080/tasks?${params.toString()}`);
      setTasks(response.data.tasks || []);
      setPaginationInfo(response.data.pagination || {
        total: 0,
        currentPage: 1,
        perPage: 9,
        totalPages: 1,
        hasMore: false
      });
    } catch (error) {
      console.error('Error fetching tasks:', error);
      setError('Failed to load tasks. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const params = new URLSearchParams();
      if (searchTerm) params.append('q', searchTerm);
      if (category) params.append('category', category);
      if (priority) params.append('priority', priority);
      if (completed) params.append('completed', completed);
      if (tag) params.append('tag', tag);
      if (dueBefore) params.append('dueBefore', new Date(dueBefore).toISOString());
      if (dueAfter) params.append('dueAfter', new Date(dueAfter).toISOString());
      
      params.append('sortBy', sortBy);
      params.append('sortDir', sortOrder);
      params.append('page', '1'); // Reset to first page on new search
      params.append('perPage', perPage.toString());

      const response = await axios.get(`http://localhost:8080/tasks?${params.toString()}`);
      setTasks(response.data.tasks || []);
      setPaginationInfo(response.data.pagination || {
        total: 0,
        currentPage: 1,
        perPage: 9,
        totalPages: 1,
        hasMore: false
      });
      setPage(1); // Reset page to 1
      setFilterDrawerOpen(false); // Close the filter drawer after search
    } catch (error) {
      console.error('Error searching tasks:', error);
      setError('Failed to search tasks. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    // Reset all filters
    setSearchTerm('');
    setCategory('');
    setPriority('');
    setTag('');
    setCompleted('');
    setDueBefore('');
    setDueAfter('');
    setShowUpcoming(false);
    setShowOverdue(false);
    setSortBy('dueDate');
    setSortOrder('asc');
    setPage(1);
    
    // Fetch tasks with reset filters
    fetchTasks();
  };

  const handleResetFilters = () => {
    setSearchTerm('');
    setCategory('');
    setPriority('');
    setTag('');
    setCompleted('');
    setDueBefore('');
    setDueAfter('');
    setShowUpcoming(false);
    setShowOverdue(false);
  };

  const handlePageChange = (event: React.ChangeEvent<unknown>, value: number) => {
    setPage(value);
  };

  const handleSelectTask = (taskId: number) => {
    setSelectedTasks(prev => {
      if (prev.includes(taskId)) {
        return prev.filter(id => id !== taskId);
      } else {
        return [...prev, taskId];
      }
    });
  };

  const handleSelectAllTasks = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      setSelectedTasks(tasks.map(task => task.id));
    } else {
      setSelectedTasks([]);
    }
  };

  const handleDeleteSelected = async () => {
    try {
      await axios.delete('http://localhost:8080/tasks/batch/delete', {
        data: { taskIds: selectedTasks }
      });
      setSnackbar({
        open: true,
        message: `${selectedTasks.length} task(s) deleted successfully`,
        severity: 'success'
      });
      setSelectedTasks([]);
      fetchTasks();
    } catch (error) {
      console.error('Error deleting tasks:', error);
      setSnackbar({
        open: true,
        message: 'Failed to delete tasks',
        severity: 'error'
      });
    }
    setConfirmDeleteOpen(false);
  };

  const handleCompleteSelected = async () => {
    try {
      await axios.post('http://localhost:8080/tasks/batch/complete', {
        taskIds: selectedTasks
      });
      setSnackbar({
        open: true,
        message: `${selectedTasks.length} task(s) marked as completed`,
        severity: 'success'
      });
      setSelectedTasks([]);
      fetchTasks();
    } catch (error) {
      console.error('Error completing tasks:', error);
      setSnackbar({
        open: true,
        message: 'Failed to complete tasks',
        severity: 'error'
      });
    }
    setConfirmCompleteOpen(false);
  };

  const handleCloseSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  const getPriorityIcon = (priority: string) => {
    switch (priority) {
      case 'high':
        return <ErrorIcon color="error" />;
      case 'medium':
        return <WarningIcon color="warning" />;
      case 'low':
        return <CheckCircleIcon color="success" />;
      default:
        return null;
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

  const getPriorityBgColor = (priority: string) => {
    switch (priority) {
      case 'high':
        return 'rgba(239, 68, 68, 0.1)';
      case 'medium':
        return 'rgba(245, 158, 11, 0.1)';
      case 'low':
        return 'rgba(16, 185, 129, 0.1)';
      default:
        return 'transparent';
    }
  };

  const renderGridView = () => (
    <Grid container spacing={3}>
      {tasks.map((task: Task) => (
        <Grid item xs={12} sm={6} md={4} key={task.id}>
          <Card
            sx={{
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              position: 'relative',
              '&:hover': {
                boxShadow: 6,
              },
            }}
          >
            {/* Checkbox for selection */}
            <Box 
              sx={{ 
                position: 'absolute', 
                top: 8, 
                left: 8, 
                zIndex: 1 
              }}
              onClick={(e) => e.stopPropagation()}
            >
              <Checkbox
                checked={selectedTasks.includes(task.id)}
                onChange={() => handleSelectTask(task.id)}
                color="primary"
              />
            </Box>
            
            <CardContent 
              sx={{ 
                flexGrow: 1, 
                pt: 4, 
                cursor: 'pointer',
                bgcolor: task.completed ? 'rgba(16, 185, 129, 0.05)' : undefined 
              }} 
              onClick={() => navigate(`/tasks/${task.id}`)}
            >
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                <Typography 
                  variant="h6" 
                  component="div"
                  sx={{ 
                    textDecoration: task.completed ? 'line-through' : 'none',
                    color: task.completed ? 'text.secondary' : 'text.primary'
                  }}
                >
                  {task.title}
                </Typography>
                <Chip
                  label={task.priority}
                  color={getPriorityColor(task.priority)}
                  size="small"
                  sx={{ fontWeight: 'bold' }}
                />
              </Box>
              
              <Typography 
                variant="body2" 
                color="text.secondary" 
                sx={{ 
                  mb: 2, 
                  display: '-webkit-box',
                  overflow: 'hidden',
                  WebkitBoxOrient: 'vertical',
                  WebkitLineClamp: 2,
                  height: '40px'
                }}
              >
                {task.description}
              </Typography>
              
              {task.category && (
                <Chip 
                  label={task.category} 
                  size="small" 
                  color="primary" 
                  variant="outlined"
                  sx={{ mb: 2 }}
                />
              )}
              
              <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', mb: 2 }}>
                {task.tags && task.tags.slice(0, 3).map((tag: string) => (
                  <Chip key={tag} label={tag} size="small" variant="outlined" />
                ))}
                {task.tags && task.tags.length > 3 && (
                  <Chip label={`+${task.tags.length - 3}`} size="small" variant="outlined" />
                )}
              </Box>
              
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 'auto' }}>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <TodayIcon sx={{ fontSize: 14, mr: 0.5, color: 'text.secondary' }} />
                  <Typography variant="caption" color="text.secondary">
                    {task.dueDate ? format(new Date(task.dueDate), 'MMM d, yyyy') : 'No due date'}
                  </Typography>
                </Box>
                
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  {task.subtasks && task.subtasks.length > 0 && (
                    <Tooltip title={`${task.subtasks.length} subtasks`}>
                      <Chip 
                        size="small" 
                        label={task.subtasks.length.toString()} 
                        variant="outlined" 
                        icon={<CheckIcon fontSize="small" />} 
                      />
                    </Tooltip>
                  )}
                  
                  {task.reminders && task.reminders.length > 0 && (
                    <Tooltip title="Has reminders">
                      <IconButton size="small">
                        <Badge badgeContent={task.reminders.length} color="secondary">
                          <CheckIcon fontSize="small" />
                        </Badge>
                      </IconButton>
                    </Tooltip>
                  )}
                  
                  <Chip
                    label={task.completed ? 'Completed' : 'Pending'}
                    color={task.completed ? 'success' : 'default'}
                    size="small"
                    variant={task.completed ? 'filled' : 'outlined'}
                  />
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      ))}
    </Grid>
  );

  const renderListView = () => (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox">
              <Checkbox
                indeterminate={selectedTasks.length > 0 && selectedTasks.length < tasks.length}
                checked={selectedTasks.length > 0 && selectedTasks.length === tasks.length}
                onChange={handleSelectAllTasks}
                color="primary"
              />
            </TableCell>
            <TableCell>Title</TableCell>
            <TableCell>Category</TableCell>
            <TableCell>Priority</TableCell>
            <TableCell>Due Date</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Actions</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {tasks.map((task: Task) => (
            <TableRow 
              key={task.id}
              sx={{ 
                '&:last-child td, &:last-child th': { border: 0 },
                bgcolor: task.completed ? 'rgba(16, 185, 129, 0.05)' : undefined,
                '&:hover': { backgroundColor: 'rgba(0, 0, 0, 0.04)' }
              }}
            >
              <TableCell padding="checkbox">
                <Checkbox
                  checked={selectedTasks.includes(task.id)}
                  onChange={() => handleSelectTask(task.id)}
                  color="primary"
                />
              </TableCell>
              <TableCell 
                component="th" 
                scope="row"
                onClick={() => navigate(`/tasks/${task.id}`)}
                sx={{ 
                  cursor: 'pointer',
                  textDecoration: task.completed ? 'line-through' : 'none',
                  color: task.completed ? 'text.secondary' : 'text.primary'
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  {getPriorityIcon(task.priority)}
                  <Typography sx={{ ml: 1 }}>{task.title}</Typography>
                </Box>
              </TableCell>
              <TableCell>{task.category || '-'}</TableCell>
              <TableCell>
                <Chip
                  label={task.priority}
                  color={getPriorityColor(task.priority)}
                  size="small"
                />
              </TableCell>
              <TableCell>
                {task.dueDate ? format(new Date(task.dueDate), 'MMM d, yyyy') : '-'}
              </TableCell>
              <TableCell>
                <Chip
                  label={task.completed ? 'Completed' : 'Pending'}
                  color={task.completed ? 'success' : 'default'}
                  size="small"
                  variant={task.completed ? 'filled' : 'outlined'}
                />
              </TableCell>
              <TableCell>
                <IconButton size="small" onClick={() => navigate(`/tasks/${task.id}/edit`)}>
                  <EditIcon fontSize="small" />
                </IconButton>
                <IconButton 
                  size="small" 
                  color="error"
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelectedTasks([task.id]);
                    setConfirmDeleteOpen(true);
                  }}
                >
                  <DeleteIcon fontSize="small" />
                </IconButton>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );

  // Render filter drawer
  const renderFilterDrawer = () => (
    <Drawer
      anchor="right"
      open={filterDrawerOpen}
      onClose={() => setFilterDrawerOpen(false)}
    >
      <Box
        sx={{ width: 300, p: 3 }}
        role="presentation"
      >
        <Typography variant="h6" gutterBottom>
          Advanced Filters
        </Typography>
        <Divider sx={{ mb: 2 }} />
        
        <Stack spacing={3}>
          <TextField
            fullWidth
            label="Search"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search tasks..."
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
          />
          
          <FormControl fullWidth>
            <InputLabel>Category</InputLabel>
            <Select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              label="Category"
            >
              <MenuItem value="">All</MenuItem>
              {categories.map((cat) => (
                <MenuItem key={cat} value={cat}>{cat}</MenuItem>
              ))}
            </Select>
          </FormControl>
          
          <FormControl fullWidth>
            <InputLabel>Priority</InputLabel>
            <Select
              value={priority}
              onChange={(e) => setPriority(e.target.value)}
              label="Priority"
            >
              <MenuItem value="">All</MenuItem>
              <MenuItem value="high">High</MenuItem>
              <MenuItem value="medium">Medium</MenuItem>
              <MenuItem value="low">Low</MenuItem>
            </Select>
          </FormControl>
          
          <FormControl fullWidth>
            <InputLabel>Status</InputLabel>
            <Select
              value={completed}
              onChange={(e) => setCompleted(e.target.value)}
              label="Status"
            >
              <MenuItem value="">All</MenuItem>
              <MenuItem value="true">Completed</MenuItem>
              <MenuItem value="false">Pending</MenuItem>
            </Select>
          </FormControl>
          
          <FormControl fullWidth>
            <InputLabel>Tag</InputLabel>
            <Select
              value={tag}
              onChange={(e) => setTag(e.target.value)}
              label="Tag"
            >
              <MenuItem value="">All</MenuItem>
              {tags.map((t) => (
                <MenuItem key={t} value={t}>{t}</MenuItem>
              ))}
            </Select>
          </FormControl>
          
          <TextField
            fullWidth
            label="Due Before"
            type="date"
            value={dueBefore}
            onChange={(e) => setDueBefore(e.target.value)}
            InputLabelProps={{ shrink: true }}
          />
          
          <TextField
            fullWidth
            label="Due After"
            type="date"
            value={dueAfter}
            onChange={(e) => setDueAfter(e.target.value)}
            InputLabelProps={{ shrink: true }}
          />
          
          <FormControlLabel
            control={
              <Switch
                checked={showUpcoming}
                onChange={(e) => {
                  setShowUpcoming(e.target.checked);
                  if (e.target.checked) setShowOverdue(false);
                }}
              />
            }
            label="Show only upcoming tasks"
          />
          
          <FormControlLabel
            control={
              <Switch
                checked={showOverdue}
                onChange={(e) => {
                  setShowOverdue(e.target.checked);
                  if (e.target.checked) setShowUpcoming(false);
                }}
              />
            }
            label="Show only overdue tasks"
          />
          
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 2 }}>
            <Button 
              variant="outlined" 
              onClick={handleResetFilters}
              startIcon={<RefreshIcon />}
            >
              Reset
            </Button>
            <Button 
              variant="contained" 
              onClick={handleSearch}
              startIcon={<SearchIcon />}
            >
              Apply Filters
            </Button>
          </Box>
        </Stack>
      </Box>
    </Drawer>
  );

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h4">Tasks</Typography>
        <Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => navigate('/tasks/new')}
            sx={{ mr: 1 }}
          >
            New Task
          </Button>
          <IconButton
            onClick={() => setFilterDrawerOpen(true)}
            color="primary"
          >
            <TuneIcon />
          </IconButton>
        </Box>
      </Box>

      {/* Action Bar */}
      <Card sx={{ mb: 4 }}>
        <CardContent>
          <Grid container spacing={2} alignItems="center">
            {/* Left side - View mode and refresh */}
            <Grid item xs={12} md={4}>
              <Box sx={{ display: 'flex', alignItems: 'center' }}>
                <Tabs 
                  value={viewMode} 
                  onChange={(e, newValue) => setViewMode(newValue)}
                  sx={{ mr: 2 }}
                >
                  <Tab
                    icon={<ViewModuleIcon />}
                    value="grid"
                    aria-label="grid view"
                  />
                  <Tab
                    icon={<ViewListIcon />}
                    value="list"
                    aria-label="list view"
                  />
                </Tabs>
                <Tooltip title="Refresh">
                  <IconButton onClick={handleRefresh}>
                    <RefreshIcon />
                  </IconButton>
                </Tooltip>
              </Box>
            </Grid>
            
            {/* Center - Batch actions */}
            <Grid item xs={12} md={4}>
              <Box sx={{ display: 'flex', justifyContent: 'center' }}>
                {selectedTasks.length > 0 && (
                  <>
                    <Button
                      variant="contained"
                      color="success"
                      startIcon={<DoneIcon />}
                      onClick={() => setConfirmCompleteOpen(true)}
                      sx={{ mr: 1 }}
                    >
                      Complete {selectedTasks.length} Tasks
                    </Button>
                    <Button
                      variant="contained"
                      color="error"
                      startIcon={<DeleteIcon />}
                      onClick={() => setConfirmDeleteOpen(true)}
                    >
                      Delete {selectedTasks.length} Tasks
                    </Button>
                  </>
                )}
              </Box>
            </Grid>
            
            {/* Right side - Sort options */}
            <Grid item xs={12} md={4}>
              <Box sx={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
                <FormControl sx={{ minWidth: 120, mr: 1 }} size="small">
                  <InputLabel>Sort By</InputLabel>
                  <Select
                    value={sortBy}
                    onChange={(e) => setSortBy(e.target.value)}
                    label="Sort By"
                  >
                    <MenuItem value="dueDate">Due Date</MenuItem>
                    <MenuItem value="priority">Priority</MenuItem>
                    <MenuItem value="title">Title</MenuItem>
                    <MenuItem value="createdAt">Created Date</MenuItem>
                  </Select>
                </FormControl>
                
                <FormControl sx={{ minWidth: 120 }} size="small">
                  <InputLabel>Order</InputLabel>
                  <Select
                    value={sortOrder}
                    onChange={(e) => setSortOrder(e.target.value)}
                    label="Order"
                  >
                    <MenuItem value="asc">Ascending</MenuItem>
                    <MenuItem value="desc">Descending</MenuItem>
                  </Select>
                </FormControl>
              </Box>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Error message */}
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {/* Loading indicator */}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
          <CircularProgress />
        </Box>
      ) : tasks.length === 0 ? (
        <Box sx={{ textAlign: 'center', my: 8 }}>
          <Typography variant="h6" color="text.secondary" gutterBottom>
            No tasks found
          </Typography>
          <Typography color="text.secondary" paragraph>
            Try changing your filters or create a new task
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => navigate('/tasks/new')}
            sx={{ mt: 2 }}
          >
            Create New Task
          </Button>
        </Box>
      ) : (
        <>
          {/* Tasks list/grid */}
          {viewMode === 'grid' ? renderGridView() : renderListView()}
          
          {/* Pagination */}
          {paginationInfo && paginationInfo.totalPages > 1 && (
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
              <Pagination
                count={paginationInfo.totalPages}
                page={paginationInfo.currentPage}
                color="primary"
                onChange={handlePageChange}
              />
            </Box>
          )}
        </>
      )}

      {/* Confirmation dialogs */}
      <Dialog open={confirmDeleteOpen} onClose={() => setConfirmDeleteOpen(false)}>
        <DialogTitle>Delete Tasks</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to delete {selectedTasks.length} task(s)? This action cannot be undone.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmDeleteOpen(false)}>Cancel</Button>
          <Button onClick={handleDeleteSelected} color="error" variant="contained">
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={confirmCompleteOpen} onClose={() => setConfirmCompleteOpen(false)}>
        <DialogTitle>Complete Tasks</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to mark {selectedTasks.length} task(s) as completed?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmCompleteOpen(false)}>Cancel</Button>
          <Button onClick={handleCompleteSelected} color="success" variant="contained">
            Complete
          </Button>
        </DialogActions>
      </Dialog>

      {/* Filter drawer */}
      {renderFilterDrawer()}

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default TaskList; 