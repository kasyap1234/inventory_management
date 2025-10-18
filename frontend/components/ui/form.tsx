import * as React from "react"
import { cn } from "@/lib/utils"
import { Label } from "./label"
import { Input } from "./input"

interface FormFieldProps {
  label: string
  name: string
  error?: string
  required?: boolean
  children: React.ReactNode
  description?: string
  className?: string
}

export function FormField({
  label,
  name,
  error,
  required,
  children,
  description,
  className,
}: FormFieldProps) {
  return (
    <div className={cn("space-y-2", className)}>
      <Label htmlFor={name} required={required}>
        {label}
      </Label>
      {children}
      {description && !error && (
        <p className="text-sm text-gray-500">{description}</p>
      )}
      {error && (
        <p className="text-sm text-red-600">{error}</p>
      )}
    </div>
  )
}

interface FormProps extends React.FormHTMLAttributes<HTMLFormElement> {
  children: React.ReactNode
}

export function Form({ children, className, ...props }: FormProps) {
  return (
    <form className={cn("space-y-6", className)} {...props}>
      {children}
    </form>
  )
}

interface FormActionsProps {
  children: React.ReactNode
  className?: string
}

export function FormActions({ children, className }: FormActionsProps) {
  return (
    <div className={cn("flex items-center gap-3 pt-4", className)}>
      {children}
    </div>
  )
}
