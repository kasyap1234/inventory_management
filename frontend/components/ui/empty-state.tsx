import { LucideIcon } from "lucide-react"
import { Button } from "./button"
import { cn } from "@/lib/utils"

interface EmptyStateProps {
  icon?: LucideIcon
  title: string
  description?: string
  action?: {
    label: string
    onClick: () => void
    variant?: "default" | "outline" | "secondary"
  }
  secondaryAction?: {
    label: string
    onClick: () => void
  }
  className?: string
  variant?: "default" | "minimal" | "gradient"
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  secondaryAction,
  className,
  variant = "default",
}: EmptyStateProps) {
  const iconWrapperStyles = {
    default: "w-20 h-20 bg-gradient-to-br from-gray-50 to-gray-100 rounded-2xl flex items-center justify-center mb-6 shadow-sm group-hover:shadow-md transition-all duration-300",
    minimal: "w-16 h-16 bg-gray-50 rounded-xl flex items-center justify-center mb-5 group-hover:bg-gray-100 transition-colors duration-200",
    gradient: "w-20 h-20 gradient-primary rounded-2xl flex items-center justify-center mb-6 shadow-colored group-hover:scale-110 transition-transform duration-300"
  }

  const iconStyles = {
    default: "h-10 w-10 text-gray-400 group-hover:text-gray-500 transition-colors duration-200",
    minimal: "h-8 w-8 text-gray-400",
    gradient: "h-10 w-10 text-white"
  }

  return (
    <div className={cn("flex flex-col items-center justify-center py-16 px-6 text-center animate-fade-in group", className)}>
      {Icon && (
        <div className={iconWrapperStyles[variant]}>
          <Icon className={iconStyles[variant]} />
        </div>
      )}
      
      <h3 className="text-xl font-bold text-gray-900 mb-3 tracking-tight">{title}</h3>
      
      {description && (
        <p className="text-base text-gray-600 mb-8 max-w-md leading-relaxed">{description}</p>
      )}
      
      {(action || secondaryAction) && (
        <div className="flex items-center gap-3">
          {action && (
            <Button 
              onClick={action.onClick}
              variant={action.variant || "default"}
              size="lg"
              className="shadow-sm hover:shadow"
            >
              {action.label}
            </Button>
          )}
          {secondaryAction && (
            <Button 
              onClick={secondaryAction.onClick}
              variant="outline"
              size="lg"
            >
              {secondaryAction.label}
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
