/**
 * Audit Logs Page
 */

import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { SearchInput } from '@/components/SearchInput';
import { Pagination } from '@/components/Pagination';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { FileText, Search, Filter } from 'lucide-react';

// Mock audit log type (replace with actual API type)
interface AuditLog {
  id: string;
  timestamp: string;
  user_id: string;
  username: string;
  action: string;
  resource: string;
  resource_id?: string;
  ip_address: string;
  user_agent: string;
  status: 'success' | 'failure';
  details?: string;
}

// Mock API function (replace with actual API call)
const fetchAuditLogs = async (): Promise<AuditLog[]> => {
  // TODO: Replace with actual API call
  // This is a placeholder that returns mock data
  return [
    {
      id: '1',
      timestamp: new Date().toISOString(),
      user_id: 'user-1',
      username: 'admin',
      action: 'login',
      resource: 'auth',
      ip_address: '192.168.1.1',
      user_agent: 'Mozilla/5.0',
      status: 'success',
    },
    {
      id: '2',
      timestamp: new Date(Date.now() - 3600000).toISOString(),
      user_id: 'user-2',
      username: 'john.doe',
      action: 'create',
      resource: 'user',
      resource_id: 'user-3',
      ip_address: '192.168.1.2',
      user_agent: 'Mozilla/5.0',
      status: 'success',
      details: 'Created new user account',
    },
  ];
};

export function AuditLogs() {
  const [searchQuery, setSearchQuery] = useState('');
  const [actionFilter, setActionFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const { data: logs, isLoading, error } = useQuery({
    queryKey: ['auditLogs'],
    queryFn: fetchAuditLogs,
  });

  // Filter logs based on search and filters
  const filteredLogs = useMemo(() => {
    if (!logs) return [];

    return logs.filter((log) => {
      const matchesSearch =
        log.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
        log.action.toLowerCase().includes(searchQuery.toLowerCase()) ||
        log.resource.toLowerCase().includes(searchQuery.toLowerCase()) ||
        log.ip_address.includes(searchQuery);

      const matchesAction = actionFilter === 'all' || log.action === actionFilter;
      const matchesStatus = statusFilter === 'all' || log.status === statusFilter;

      return matchesSearch && matchesAction && matchesStatus;
    });
  }, [logs, searchQuery, actionFilter, statusFilter]);

  // Paginate filtered logs
  const paginatedLogs = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return filteredLogs.slice(start, end);
  }, [filteredLogs, currentPage, pageSize]);

  const totalPages = Math.ceil(filteredLogs.length / pageSize);

  // Get unique actions for filter
  const uniqueActions = useMemo(() => {
    if (!logs) return [];
    return Array.from(new Set(logs.map((log) => log.action))).sort();
  }, [logs]);

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-3xl font-bold">Audit Logs</h1>
        <div className="text-center py-8">Loading audit logs...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <h1 className="text-3xl font-bold">Audit Logs</h1>
        <div className="p-4 text-red-600">
          Error loading audit logs: {error instanceof Error ? error.message : 'Unknown error'}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Audit Logs</h1>
          <p className="text-gray-600 mt-1">View and search system audit logs</p>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="space-y-2">
              <Label>Search</Label>
              <SearchInput
                value={searchQuery}
                onChange={setSearchQuery}
                placeholder="Search by user, action, resource, or IP..."
              />
            </div>

            <div className="space-y-2">
              <Label>Action</Label>
              <Select value={actionFilter} onValueChange={setActionFilter}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Actions</SelectItem>
                  {uniqueActions.map((action) => (
                    <SelectItem key={action} value={action}>
                      {action}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Status</Label>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="success">Success</SelectItem>
                  <SelectItem value="failure">Failure</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-end">
              <Button
                variant="outline"
                onClick={() => {
                  setSearchQuery('');
                  setActionFilter('all');
                  setStatusFilter('all');
                  setCurrentPage(1);
                }}
                className="w-full"
              >
                Clear Filters
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Audit Logs Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Log Entries
          </CardTitle>
          <CardDescription>
            {filteredLogs.length} log entry(s) found
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>User</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Resource</TableHead>
                  <TableHead>IP Address</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Details</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {paginatedLogs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="font-mono text-xs">
                      {new Date(log.timestamp).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <div>
                        <div className="font-medium">{log.username}</div>
                        <div className="text-xs text-gray-500">{log.user_id}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs">
                        {log.action}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div>
                        <div className="font-medium">{log.resource}</div>
                        {log.resource_id && (
                          <div className="text-xs text-gray-500">{log.resource_id}</div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{log.ip_address}</TableCell>
                    <TableCell>
                      <span
                        className={`px-2 py-1 rounded text-xs ${
                          log.status === 'success'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-red-100 text-red-800'
                        }`}
                      >
                        {log.status}
                      </span>
                    </TableCell>
                    <TableCell className="text-xs text-gray-600 max-w-xs truncate">
                      {log.details || '-'}
                    </TableCell>
                  </TableRow>
                ))}
                {paginatedLogs.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-gray-500 py-8">
                      {filteredLogs.length === 0 && logs && logs.length > 0
                        ? 'No logs match your filters'
                        : 'No audit logs found'}
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>

            {filteredLogs.length > 0 && (
              <Pagination
                currentPage={currentPage}
                totalPages={totalPages}
                pageSize={pageSize}
                totalItems={filteredLogs.length}
                onPageChange={setCurrentPage}
                onPageSizeChange={(size) => {
                  setPageSize(size);
                  setCurrentPage(1);
                }}
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Info Card */}
      <Card className="bg-blue-50 border-blue-200">
        <CardContent className="pt-6">
          <div className="flex items-start gap-3">
            <FileText className="h-5 w-5 text-blue-600 mt-0.5" />
            <div>
              <h3 className="font-semibold text-blue-900 mb-1">About Audit Logs</h3>
              <p className="text-sm text-blue-800">
                Audit logs record all system activities including user actions, authentication
                events, and administrative changes. Use filters to find specific events or search
                by user, action, or resource.
              </p>
              <p className="text-xs text-blue-700 mt-2">
                <strong>Note:</strong> This is a placeholder implementation. Connect to your actual
                audit log API endpoint to display real data.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

