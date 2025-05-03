import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  Divider,
  Button,
  Alert,
  Paper,
  CardHeader,
  Tooltip,
  IconButton,
  Avatar,
  LinearProgress,
  Chip,
} from '@mui/material';
import {
  CheckCircle as CheckCircleIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  Category as CategoryIcon,
  Refresh as RefreshIcon,
  Timer as TimerIcon,
  Assignment as AssignmentIcon,
  CalendarMonth as CalendarIcon,
  Alarm as AlarmIcon,
  Add as AddIcon,
  MoreVert as MoreVertIcon,
  Flag as FlagIcon,
  TaskAlt as TaskAltIcon,
} from '@mui/icons-material';
import axios from 'axios';
import { format, isAfter, parseISO } from 'date-fns';

interface TaskStats {
  total: number;
  completed: number;
  pending: number;
  overdue: number;
  byCategory: Record<string, number>;
  byPriority: {
    low: number;
    medium: number;
    high: number;
  };
}

interface TaskAnalytics {
  total_tasks: number;
  completed_tasks: number;
  completion_rate: number;
  category_distribution: Record<string, number>;
  priority_distribution: Record<string, number>;
  avg_completion_time: number;
  top_tags: Array<{ tag: string; count: number }>;
  overdue_tasks: number;
  weekly_activity: Record<string, number>;
}

interface RecentTask {
  id: number;
  title: string;
  description: string;
  dueDate: string;
  priority: 'low' | 'medium' | 'high';
  completed: boolean;
  category: string;
  tags: string[];
}

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<TaskStats | null>(null);
  const [analytics, setAnalytics] = useState<TaskAnalytics | null>(null);
  const [recentTasks, setRecentTasks] = useState<RecentTask[]>([]);
  const [upcomingTasks, setUpcomingTasks] = useState<RecentTask[]>([]);
  const [overdueTasks, setOverdueTasks] = useState<RecentTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      // Get stats - using the correct endpoint
      const statsResponse = await axios.get('http://localhost:8080/api/v1/stats');
      
      // Create a default stats object if some properties are missing
      const statsData: TaskStats = {
        total: statsResponse.data.total || 0,
        completed: statsResponse.data.completed || 0,
        pending: statsResponse.data.pending || 0,
        overdue: statsResponse.data.overdue || 0,
        byCategory: statsResponse.data.byCategory || {},
        byPriority: {
          low: statsResponse.data.byPriority?.low || 0,
          medium: statsResponse.data.byPriority?.medium || 0,
          high: statsResponse.data.byPriority?.high || 0,
        }
      };
      
      setStats(statsData);
      
      // Get analytics data from our new endpoint
      try {
        const analyticsResponse = await axios.get('http://localhost:8080/api/v1/analytics');
        setAnalytics(analyticsResponse.data);
      } catch (e) {
        console.error('Analytics endpoint not available:', e);
        // Fall back to stats data
        setAnalytics({
          total_tasks: statsData.total,
          completed_tasks: statsData.completed,
          completion_rate: statsData.total > 0 ? (statsData.completed / statsData.total) * 100 : 0,
          category_distribution: statsData.byCategory,
          priority_distribution: statsData.byPriority,
          avg_completion_time: 0,
          top_tags: [],
          overdue_tasks: statsData.overdue,
          weekly_activity: {},
        });
      }
      
      // Get tasks
      const tasksResponse = await axios.get('http://localhost:8080/api/v1/tasks?perPage=20');
      
      if (tasksResponse.data.tasks) {
        const allTasks = tasksResponse.data.tasks;
        
        // Recent tasks (last 5 created)
        const sortedByDate = [...allTasks].sort((a, b) => 
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
        setRecentTasks(sortedByDate.slice(0, 5));
        
        // Upcoming tasks (next 5 due, not completed and not overdue)
        const now = new Date();
        const upcoming = allTasks
          .filter((task: RecentTask) => !task.completed && task.dueDate && isAfter(parseISO(task.dueDate), now))
          .sort((a: RecentTask, b: RecentTask) => parseISO(a.dueDate).getTime() - parseISO(b.dueDate).getTime());
        setUpcomingTasks(upcoming.slice(0, 5));
        
        // Overdue tasks
        const overdue = allTasks
          .filter((task: RecentTask) => !task.completed && task.dueDate && !isAfter(parseISO(task.dueDate), now))
          .sort((a: RecentTask, b: RecentTask) => parseISO(b.dueDate).getTime() - parseISO(a.dueDate).getTime());
        setOverdueTasks(overdue.slice(0, 5));
      } else {
        setRecentTasks([]);
        setUpcomingTasks([]);
        setOverdueTasks([]);
      }
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
      setError('Failed to load dashboard data. Please try again.');
    } finally {
      setLoading(false);
    }
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
        return 'primary';
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box>
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
        <Button 
          variant="contained" 
          startIcon={<RefreshIcon />}
          onClick={fetchDashboardData}
        >
          Retry
        </Button>
      </Box>
    );
  }

  const completionRate = analytics?.completion_rate ?? 0;

  return (
    <Box>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="h4">Task Dashboard</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => navigate('/tasks/new')}
        >
          New Task
        </Button>
      </Box>

      {/* Task Statistics Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
              <Box sx={{ position: 'relative', display: 'inline-flex', mb: 1 }}>
                <CircularProgress 
                  variant="determinate" 
                  value={completionRate} 
                  size={80} 
                  thickness={5} 
                  color="primary"
                />
                <Box
                  sx={{
                    top: 0,
                    left: 0,
                    bottom: 0,
                    right: 0,
                    position: 'absolute',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Typography variant="h6" component="div" color="primary">
                    {`${Math.round(completionRate)}%`}
                  </Typography>
                </Box>
              </Box>
              <Typography variant="h6">Completion Rate</Typography>
              <Typography variant="body2" color="text.secondary">
                {analytics?.completed_tasks} of {analytics?.total_tasks} tasks completed
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ display: 'flex', alignItems: 'center', height: '100%' }}>
              <Avatar sx={{ bgcolor: 'warning.main', mr: 2 }}>
                <TimerIcon />
              </Avatar>
              <Box>
                <Typography variant="h6" gutterBottom>
                  Average Time
                </Typography>
                <Typography variant="h5">
                  {analytics?.avg_completion_time ? 
                    `${Math.round(analytics.avg_completion_time)} hours` : 
                    'N/A'}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  to complete tasks
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ display: 'flex', alignItems: 'center', height: '100%' }}>
              <Avatar sx={{ bgcolor: 'success.main', mr: 2 }}>
                <TaskAltIcon />
              </Avatar>
              <Box>
                <Typography variant="h6" gutterBottom>
                  Pending Tasks
                </Typography>
                <Typography variant="h5">
                  {stats?.pending || 0}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  tasks to be completed
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent sx={{ display: 'flex', alignItems: 'center', height: '100%' }}>
              <Avatar sx={{ bgcolor: 'error.main', mr: 2 }}>
                <AlarmIcon />
              </Avatar>
              <Box>
                <Typography variant="h6" gutterBottom>
                  Overdue Tasks
                </Typography>
                <Typography variant="h5">
                  {analytics?.overdue_tasks || 0}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  need immediate attention
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        {/* Tasks by Priority */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardHeader 
              title="Priority Distribution" 
              titleTypographyProps={{ variant: 'h6' }}
              action={
                <Tooltip title="View all tasks">
                  <IconButton onClick={() => navigate('/tasks')}>
                    <MoreVertIcon />
                  </IconButton>
                </Tooltip>
              }
            />
            <CardContent>
              <Box sx={{ mt: 2 }}>
                <Typography variant="body2" gutterBottom>
                  High Priority ({analytics?.priority_distribution?.high || 0})
                </Typography>
                <LinearProgress 
                  variant="determinate" 
                  value={analytics?.total_tasks ? (analytics.priority_distribution?.high / analytics.total_tasks) * 100 : 0} 
                  color="error" 
                  sx={{ height: 10, borderRadius: 5, mb: 2 }}
                />
                
                <Typography variant="body2" gutterBottom>
                  Medium Priority ({analytics?.priority_distribution?.medium || 0})
                </Typography>
                <LinearProgress 
                  variant="determinate" 
                  value={analytics?.total_tasks ? (analytics.priority_distribution?.medium / analytics.total_tasks) * 100 : 0} 
                  color="warning" 
                  sx={{ height: 10, borderRadius: 5, mb: 2 }}
                />
                
                <Typography variant="body2" gutterBottom>
                  Low Priority ({analytics?.priority_distribution?.low || 0})
                </Typography>
                <LinearProgress 
                  variant="determinate" 
                  value={analytics?.total_tasks ? (analytics.priority_distribution?.low / analytics.total_tasks) * 100 : 0} 
                  color="success" 
                  sx={{ height: 10, borderRadius: 5 }}
                />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Tasks by Category */}
        <Grid item xs={12} md={4}>
          <Card sx={{ height: '100%' }}>
            <CardHeader 
              title="Categories" 
              titleTypographyProps={{ variant: 'h6' }}
              action={
                <Tooltip title="View categories">
                  <IconButton onClick={() => navigate('/tasks')}>
                    <MoreVertIcon />
                  </IconButton>
                </Tooltip>
              }
            />
            <CardContent sx={{ overflowY: 'auto', maxHeight: 300 }}>
              {Object.keys(analytics?.category_distribution || {}).length > 0 ? (
                <List>
                  {Object.entries(analytics?.category_distribution || {}).map(([category, count]) => (
                    <React.Fragment key={category}>
                      <ListItem disablePadding sx={{ py: 1 }}>
                        <CategoryIcon sx={{ mr: 2, color: 'primary.main' }} />
                        <ListItemText
                          primary={category || 'Uncategorized'}
                          secondary={`${count} tasks`}
                        />
                      </ListItem>
                      <Divider />
                    </React.Fragment>
                  ))}
                </List>
              ) : (
                <Typography color="text.secondary">No categories available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Popular Tags */}
        <Grid item xs={12} md={4}>
          <Card sx={{ height: '100%' }}>
            <CardHeader 
              title="Popular Tags" 
              titleTypographyProps={{ variant: 'h6' }}
              action={
                <Tooltip title="View all tags">
                  <IconButton onClick={() => navigate('/tasks')}>
                    <MoreVertIcon />
                  </IconButton>
                </Tooltip>
              }
            />
            <CardContent>
              {analytics?.top_tags && analytics.top_tags.length > 0 ? (
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                  {analytics.top_tags.map((tagData) => (
                    <Tooltip key={tagData.tag} title={`${tagData.count} tasks`}>
                      <Chip
                        label={tagData.tag}
                        size="small"
                        color="primary"
                        variant="outlined"
                        onClick={() => navigate(`/tasks?tag=${tagData.tag}`)}
                      />
                    </Tooltip>
                  ))}
                </Box>
              ) : (
                <Typography color="text.secondary">No tags available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Upcoming Tasks */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardHeader 
              title="Upcoming Tasks" 
              titleTypographyProps={{ variant: 'h6' }}
              avatar={<Avatar sx={{ bgcolor: 'primary.main' }}><CalendarIcon /></Avatar>}
              action={
                <Button 
                  variant="outlined" 
                  size="small" 
                  onClick={() => navigate('/tasks')}
                  sx={{ borderRadius: '20px' }}
                >
                  View All
                </Button>
              }
            />
            <Divider />
            <CardContent sx={{ p: 0 }}>
              {upcomingTasks.length > 0 ? (
                <List>
                  {upcomingTasks.map((task) => (
                    <React.Fragment key={task.id}>
                      <ListItem 
                        button 
                        onClick={() => navigate(`/tasks/${task.id}`)}
                        sx={{ pl: 3 }}
                      >
                        <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                          <FlagIcon color={getPriorityColor(task.priority)} sx={{ mr: 2 }} />
                          <ListItemText
                            primary={task.title}
                            secondary={
                              <React.Fragment>
                                <Typography component="span" variant="body2" color="text.primary">
                                  {task.category || 'Uncategorized'}
                                </Typography>
                                {' — '}
                                {task.dueDate ? `Due: ${format(new Date(task.dueDate), 'MMM d, yyyy')}` : 'No due date'}
                              </React.Fragment>
                            }
                          />
                        </Box>
                      </ListItem>
                      <Divider />
                    </React.Fragment>
                  ))}
                </List>
              ) : (
                <Box sx={{ p: 3, textAlign: 'center' }}>
                  <Typography color="text.secondary">No upcoming tasks</Typography>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Overdue Tasks */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardHeader 
              title="Overdue Tasks" 
              titleTypographyProps={{ variant: 'h6' }}
              avatar={<Avatar sx={{ bgcolor: 'error.main' }}><AlarmIcon /></Avatar>}
              action={
                <Button 
                  variant="outlined" 
                  color="error" 
                  size="small" 
                  onClick={() => navigate('/tasks')}
                  sx={{ borderRadius: '20px' }}
                >
                  View All
                </Button>
              }
            />
            <Divider />
            <CardContent sx={{ p: 0 }}>
              {overdueTasks.length > 0 ? (
                <List>
                  {overdueTasks.map((task) => (
                    <React.Fragment key={task.id}>
                      <ListItem 
                        button 
                        onClick={() => navigate(`/tasks/${task.id}`)}
                        sx={{ pl: 3, bgcolor: 'error.lighter' }}
                      >
                        <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                          <FlagIcon color={getPriorityColor(task.priority)} sx={{ mr: 2 }} />
                          <ListItemText
                            primary={task.title}
                            secondary={
                              <React.Fragment>
                                <Typography component="span" variant="body2" color="text.primary">
                                  {task.category || 'Uncategorized'}
                                </Typography>
                                {' — '}
                                <Typography component="span" variant="body2" color="error">
                                  {task.dueDate ? `Due: ${format(new Date(task.dueDate), 'MMM d, yyyy')}` : 'No due date'}
                                </Typography>
                              </React.Fragment>
                            }
                          />
                        </Box>
                      </ListItem>
                      <Divider />
                    </React.Fragment>
                  ))}
                </List>
              ) : (
                <Box sx={{ p: 3, textAlign: 'center' }}>
                  <Typography color="text.secondary">No overdue tasks</Typography>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard; 