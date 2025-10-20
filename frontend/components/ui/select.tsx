'use client'

import * as React from "react"

interface SelectProps {
  value?: string
  onValueChange?: (value: string) => void
  children?: React.ReactNode
}

interface SelectTriggerProps {
  className?: string
  children?: React.ReactNode
}

interface SelectContentProps {
  children?: React.ReactNode
}

interface SelectItemProps {
  value: string
  children?: React.ReactNode
}

interface SelectValueProps {
  placeholder?: string
}

const SelectContext = React.createContext<{
  value?: string
  onValueChange?: (value: string) => void
  open: boolean
  setOpen: (open: boolean) => void
  selectedLabel?: string
  setSelectedLabel?: (label: string) => void
}>({ open: false, setOpen: () => {} })

export function Select({ value, onValueChange, children }: SelectProps) {
  const [open, setOpen] = React.useState(false)
  const [selectedLabel, setSelectedLabel] = React.useState<string>()

  return (
    <SelectContext.Provider value={{ value, onValueChange, open, setOpen, selectedLabel, setSelectedLabel }}>
      {children}
    </SelectContext.Provider>
  )
}

export function SelectTrigger({ className, children }: SelectTriggerProps) {
  const { setOpen, open } = React.useContext(SelectContext)

  return (
    <button
      type="button"
      onClick={() => setOpen(!open)}
      className={`flex h-9 w-full items-center justify-between rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50 ${className || ''}`}
    >
      {children}
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4 opacity-50">
        <path d="m6 9 6 6 6-6" />
      </svg>
    </button>
  )
}

export function SelectContent({ children }: SelectContentProps) {
  const { open } = React.useContext(SelectContext)

  if (!open) return null

  return (
    <div className="relative z-50">
      <div className="absolute top-1 w-full rounded-md border bg-popover p-1 text-popover-foreground shadow-md">
        {children}
      </div>
    </div>
  )
}

export function SelectItem({ value, children }: SelectItemProps) {
  const { onValueChange, setOpen, value: selectedValue, setSelectedLabel } = React.useContext(SelectContext)

  return (
    <div
      className={`relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-2 pr-8 text-sm outline-none hover:bg-accent hover:text-accent-foreground ${selectedValue === value ? 'bg-accent text-accent-foreground' : ''}`}
      onClick={() => {
        onValueChange?.(value)
        setSelectedLabel?.(children as string)
        setOpen(false)
      }}
    >
      {children}
    </div>
  )
}

export function SelectValue({ placeholder }: SelectValueProps) {
  const { value, selectedLabel } = React.useContext(SelectContext)

  return <span>{selectedLabel || value || placeholder || 'Select...'}</span>
}