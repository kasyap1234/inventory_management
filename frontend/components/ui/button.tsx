import * as React from "react"
import { cn } from "@/lib/utils"

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'default', ...props }, ref) => {
    const baseStyles = "inline-flex items-center justify-center rounded-xl font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/20 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98]"
    
    const variants = {
      default: "bg-gray-900 text-white hover:bg-gray-800 shadow-sm hover:shadow hover:scale-[1.01]",
      destructive: "bg-red-600 text-white hover:bg-red-700 shadow-sm hover:shadow hover:scale-[1.01]",
      outline: "border border-gray-200 bg-white hover:bg-gray-50 hover:border-gray-300 text-gray-900",
      secondary: "bg-gray-50 text-gray-900 hover:bg-gray-100 hover:scale-[1.01]",
      ghost: "hover:bg-gray-50 text-gray-700 hover:text-gray-900",
      link: "text-blue-600 underline-offset-4 hover:underline hover:text-blue-700",
    }
    
    const sizes = {
      default: "h-11 px-5 py-2.5 text-sm",
      sm: "h-9 px-4 text-sm",
      lg: "h-12 px-8 text-base",
      icon: "h-10 w-10",
    }
    
    return (
      <button
        className={cn(
          baseStyles,
          variants[variant],
          sizes[size],
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }
