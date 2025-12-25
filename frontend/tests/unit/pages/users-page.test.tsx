import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import React from 'react';

// Define mocks using vi.hoisted() to ensure they're available when vi.mock is hoisted
const { mockUserService, mockInvitationService, mockRoleService } = vi.hoisted(() => ({
  mockUserService: {
    list: vi.fn(),
  },
  mockInvitationService: {
    list: vi.fn(),
    create: vi.fn(),
    revoke: vi.fn(),
  },
  mockRoleService: {
    list: vi.fn(),
  },
}));

vi.mock('@/lib/services', () => ({
  userService: mockUserService,
  invitationService: mockInvitationService,
  roleService: mockRoleService,
}));

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({ user: { id: 'user-1' } }),
}));

// Import component after mocks are set up
import UsersPage from '@/app/users/page';

describe('UsersPage invitations', () => {
  beforeEach(() => {
    mockUserService.list.mockResolvedValue({ data: { users: [] } });
    mockInvitationService.list.mockResolvedValue({ data: { invitations: [] } });
    mockInvitationService.create.mockResolvedValue({ data: {} });
    mockRoleService.list.mockResolvedValue({
      data: { roles: [{ id: 'role-1', name: 'User' }] },
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('sends invitation with selected role', async () => {
    render(<UsersPage />);

    await waitFor(() => {
      expect(mockInvitationService.list).toHaveBeenCalled();
    });

    const emailInput = screen.getByLabelText(/Email Address/i);
    fireEvent.change(emailInput, { target: { value: 'invitee@example.com' } });

    const roleSelect = screen.getByLabelText(/Role/i);
    fireEvent.change(roleSelect, { target: { value: 'role-1' } });

    const submitButton = screen.getByRole('button', { name: /Send Invitation/i });
    fireEvent.click(submitButton);

    await waitFor(() =>
      expect(mockInvitationService.create).toHaveBeenCalledWith({
        email: 'invitee@example.com',
        role_id: 'role-1',
      }),
    );
  });
});
