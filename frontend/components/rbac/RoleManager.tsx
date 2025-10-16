'use client'

import React, { useState } from 'react'
import { Plus, Edit, Trash2, Users, Shield, CheckCircle, XCircle } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { roleAPI } from '@/lib/api'
import { toast } from 'react-hot-toast'

interface Role {
  id: string
  name: string
  description: string
  is_active: boolean
  user_count?: number
  permission_count?: number
  created_at: string
}

interface RoleFormData {
  name: string
  description: string
  is_active: boolean
}

export default function RoleManager() {
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [formData, setFormData] = useState<RoleFormData>({
    name: '',
    description: '',
    is_active: true,
  })

  const queryClient = useQueryClient()

  // Fetch roles
  const { data: roles = [], isLoading } = useQuery({
    queryKey: ['roles'],
    queryFn: async () => {
      const response = await roleAPI.list()
      return (response.data as any)?.roles ?? []
    }
  })

  // Save role mutation
  const saveRoleMutation = useMutation({
    mutationFn: async (roleData: RoleFormData) => {
      if (editingRole) {
        await roleAPI.update(editingRole.id, { name: roleData.name, description: roleData.description })
      } else {
        await roleAPI.create({ name: roleData.name, description: roleData.description })
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      toast.success(`Role ${editingRole ? 'updated' : 'created'} successfully`)
      resetForm()
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to save role')
    }
  })

  // Delete role mutation
  const deleteRoleMutation = useMutation({
    mutationFn: async (roleId: string) => {
      await roleAPI.delete(roleId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      toast.success('Role deleted successfully')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to delete role')
    }
  })

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      is_active: true,
    })
    setEditingRole(null)
    setIsFormOpen(false)
  }

  const handleEdit = (role: Role) => {
    setEditingRole(role)
    setFormData({
      name: role.name,
      description: role.description,
      is_active: role.is_active,
    })
    setIsFormOpen(true)
  }

  const handleSave = () => {
    if (!formData.name.trim()) {
      toast.error('Role name is required')
      return
    }

    saveRoleMutation.mutate(formData)
  }

  const handleDelete = (role: Role) => {
    if (role.user_count && role.user_count > 0) {
      toast.error(`Cannot delete role: ${role.user_count} users are assigned to this role`)
      return
    }

    if (window.confirm(`Are you sure you want to delete the role "${role.name}"?`)) {
      deleteRoleMutation.mutate(role.id)
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Role Management</h1>
          <p className="text-gray-600">Create and manage user roles and permissions</p>
        </div>
        <button
          onClick={() => setIsFormOpen(true)}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          <span>Create Role</span>
        </button>
      </div>

      {/* Roles Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {isLoading ? (
          <div className="col-span-full text-center py-12 text-gray-500">
            Loading roles...
          </div>
        ) : roles.length === 0 ? (
          <div className="col-span-full text-center py-12">
            <Shield className="w-16 h-16 mx-auto text-gray-300 mb-4" />
            <p className="text-gray-500">No roles configured</p>
            <button
              onClick={() => setIsFormOpen(true)}
              className="mt-4 text-blue-600 hover:text-blue-800"
            >
              Create your first role
            </button>
          </div>
        ) : (
          roles.map((role: Role) => (
            <div
              key={role.id}
              className="bg-white rounded-lg shadow border border-gray-200 p-6 hover:shadow-lg transition-shadow"
            >
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center space-x-3">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <Shield className="w-6 h-6 text-blue-600" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">{role.name}</h3>
                    <div className="flex items-center mt-1">
                      {role.is_active ? (
                        <span className="flex items-center text-xs text-green-600">
                          <CheckCircle className="w-3 h-3 mr-1" />
                          Active
                        </span>
                      ) : (
                        <span className="flex items-center text-xs text-gray-400">
                          <XCircle className="w-3 h-3 mr-1" />
                          Inactive
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                {role.description || 'No description provided'}
              </p>

              <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                <div className="flex items-center">
                  <Users className="w-4 h-4 mr-1" />
                  <span>{role.user_count || 0} users</span>
                </div>
                <div className="flex items-center">
                  <Shield className="w-4 h-4 mr-1" />
                  <span>{role.permission_count || 0} permissions</span>
                </div>
              </div>

              <div className="flex items-center space-x-2 pt-4 border-t border-gray-200">
                <button
                  onClick={() => handleEdit(role)}
                  className="flex-1 flex items-center justify-center space-x-1 px-3 py-2 text-gray-700 bg-gray-100 rounded hover:bg-gray-200"
                >
                  <Edit className="w-4 h-4" />
                  <span>Edit</span>
                </button>
                
                <button
                  onClick={() => window.location.href = `/roles/${role.id}/permissions`}
                  className="flex-1 flex items-center justify-center space-x-1 px-3 py-2 text-blue-700 bg-blue-100 rounded hover:bg-blue-200"
                >
                  <Shield className="w-4 h-4" />
                  <span>Permissions</span>
                </button>
                
                <button
                  onClick={() => handleDelete(role)}
                  disabled={deleteRoleMutation.isPending || (role.user_count || 0) > 0}
                  className="px-3 py-2 text-red-600 bg-red-100 rounded hover:bg-red-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Role Form Modal */}
      {isFormOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-md">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-gray-900">
                  {editingRole ? 'Edit' : 'Create'} Role
                </h2>
                <button
                  onClick={resetForm}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <XCircle className="w-6 h-6" />
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Role Name
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Enter role name"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Description
                  </label>
                  <textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Enter role description"
                  />
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="is_active"
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                  <label htmlFor="is_active" className="ml-2 text-sm text-gray-700">
                    Active
                  </label>
                </div>

                <div className="flex justify-end space-x-4 pt-6 border-t">
                  <button
                    onClick={resetForm}
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleSave}
                    disabled={saveRoleMutation.isPending}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                  >
                    {editingRole ? 'Update' : 'Create'} Role
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}