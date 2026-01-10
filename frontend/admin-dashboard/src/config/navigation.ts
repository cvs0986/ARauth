/**
 * Navigation Configuration
 * 
 * GUARDRAIL #1: Backend Is Law
 * - Permission requirements match backend permission model
 * 
 * GUARDRAIL #6: UI Quality Bar
 * - Professional IAM vendor navigation structure
 * - Clear hierarchy and grouping
 */

import type { LucideIcon } from 'lucide-react';
import {
    LayoutDashboard,
    Building2,
    Users,
    Shield,
    Key,
    Settings,
    FileText,
    Lock,
    Webhook,
    Globe,
    Link2,
    Zap,
    Activity,
} from 'lucide-react';

export interface NavigationItem {
    name: string;
    href: string;
    icon: LucideIcon;
    permission?: string;
    badge?: string;
}

export interface NavigationGroup {
    name: string;
    items: NavigationItem[];
}

/**
 * SYSTEM Mode Navigation
 * For platform administrators managing all tenants
 */
export const systemNavigation: NavigationGroup[] = [
    {
        name: 'Overview',
        items: [
            {
                name: 'Dashboard',
                href: '/',
                icon: LayoutDashboard,
            },
        ],
    },
    {
        name: 'Platform',
        items: [
            {
                name: 'Tenants',
                href: '/tenants',
                icon: Building2,
                permission: 'tenant:read',
            },
            {
                name: 'System Users',
                href: '/users',
                icon: Users,
                permission: 'users:read',
            },
            {
                name: 'System Roles',
                href: '/roles',
                icon: Shield,
                permission: 'roles:read',
            },
            {
                name: 'Capabilities',
                href: '/capabilities/system',
                icon: Zap,
                permission: 'system:configure',
            },
        ],
    },
    {
        name: 'Security',
        items: [
            {
                name: 'MFA Management',
                href: '/mfa',
                icon: Lock,
            },
            {
                name: 'Audit Logs',
                href: '/audit',
                icon: FileText,
                permission: 'audit:read',
            },
        ],
    },
    {
        name: 'Configuration',
        items: [
            {
                name: 'Settings',
                href: '/settings',
                icon: Settings,
                permission: 'system:configure',
            },
        ],
    },
];

/**
 * TENANT Mode Navigation
 * For tenant administrators or SYSTEM users viewing a specific tenant
 */
export const tenantNavigation: NavigationGroup[] = [
    {
        name: 'Overview',
        items: [
            {
                name: 'Dashboard',
                href: '/',
                icon: LayoutDashboard,
            },
        ],
    },
    {
        name: 'Identity',
        items: [
            {
                name: 'Users',
                href: '/users',
                icon: Users,
                permission: 'users:read',
            },
            {
                name: 'Roles',
                href: '/roles',
                icon: Shield,
                permission: 'roles:read',
            },
            {
                name: 'Permissions',
                href: '/permissions',
                icon: Key,
                permission: 'permissions:read',
            },
        ],
    },
    {
        name: 'Access',
        items: [
            {
                name: 'OAuth2 Clients',
                href: '/oauth/clients',
                icon: Key,
                permission: 'oauth:clients:read',
                badge: 'Soon',
            },
            {
                name: 'SCIM Provisioning',
                href: '/scim',
                icon: Link2,
                permission: 'scim:read',
                badge: 'Soon',
            },
        ],
    },
    {
        name: 'Federation',
        items: [
            {
                name: 'External IdPs',
                href: '/federation/idps',
                icon: Globe,
                permission: 'federation:idp:read',
                badge: 'Soon',
            },
            {
                name: 'Identity Linking',
                href: '/federation/linking',
                icon: Link2,
                permission: 'federation:link',
                badge: 'Soon',
            },
        ],
    },
    {
        name: 'Security',
        items: [
            {
                name: 'MFA Settings',
                href: '/mfa',
                icon: Lock,
            },
            {
                name: 'Active Sessions',
                href: '/security/sessions',
                icon: Activity,
                badge: 'Soon',
            },
            {
                name: 'Audit Logs',
                href: '/audit',
                icon: FileText,
                permission: 'audit:read',
            },
        ],
    },
    {
        name: 'Advanced',
        items: [
            {
                name: 'Webhooks',
                href: '/webhooks',
                icon: Webhook,
                permission: 'webhooks:read',
                badge: 'Soon',
            },
            {
                name: 'Settings',
                href: '/settings',
                icon: Settings,
                permission: 'settings:read',
            },
        ],
    },
];
