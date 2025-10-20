'use client'

import * as React from "react"

interface TabsProps {
  value?: string
  onValueChange?: (value: string) => void
  defaultValue?: string
  children?: React.ReactNode
  className?: string
}

interface TabsListProps {
  className?: string
  children?: React.ReactNode
}

interface TabsTriggerProps {
  value: string
  className?: string
  children?: React.ReactNode
}

interface TabsContentProps {
  value: string
  className?: string
  children?: React.ReactNode
}

const TabsContext = React.createContext<{
  value: string
  onValueChange: (value: string) => void
}>({ value: '', onValueChange: () => {} })

export function Tabs({ value, onValueChange, defaultValue, children, className }: TabsProps) {
  const [selectedValue, setSelectedValue] = React.useState(value || defaultValue || '')

  React.useEffect(() => {
    if (value !== undefined) {
      setSelectedValue(value)
    }
  }, [value])

  const handleValueChange = (newValue: string) => {
    setSelectedValue(newValue)
    onValueChange?.(newValue)
  }

  return (
    <TabsContext.Provider value={{ value: selectedValue, onValueChange: handleValueChange }}>
      <div className={className}>
        {children}
      </div>
    </TabsContext.Provider>
  )
}

export function TabsList({ className, children }: TabsListProps) {
  return (
    <div className={`inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground ${className || ''}`}>
      {children}
    </div>
  )
}

export function TabsTrigger({ value, className, children }: TabsTriggerProps) {
  const { value: selectedValue, onValueChange } = React.useContext(TabsContext)
  const isSelected = selectedValue === value

  return (
    <button
      type="button"
      onClick={() => onValueChange(value)}
      className={`inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 ${
        isSelected
          ? 'bg-background text-foreground shadow'
          : 'hover:bg-gray-100'
      } ${className || ''}`}
    >
      {children}
    </button>
  )
}

export function TabsContent({ value, className, children }: TabsContentProps) {
  const { value: selectedValue } = React.useContext(TabsContext)

  if (selectedValue !== value) return null

  return (
    <div className={`mt-2 ${className || ''}`}>
      {children}
    </div>
  )
}