import { useState, type FormEvent } from 'react'
import { Card, CardTitle, CardContent } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Select } from '@/components/ui/Select'
import { Button } from '@/components/ui/Button'
import { sanitize } from '@/lib/utils'
import { Save, Palette, Store, Mail, Globe } from 'lucide-react'

interface StoreSettings {
  name: string
  description: string
  email: string
  phone: string
  whatsapp: string
  address: string
  primaryColor: string
  template: string
  logo: string
  instagram: string
  facebook: string
}

const defaultSettings: StoreSettings = {
  name: '',
  description: '',
  email: '',
  phone: '',
  whatsapp: '',
  address: '',
  primaryColor: '#6366f1',
  template: 'modern',
  logo: '',
  instagram: '',
  facebook: '',
}

const STORAGE_KEY = 'sm_store_settings'

function loadSettings(): StoreSettings {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    return saved ? { ...defaultSettings, ...JSON.parse(saved) } : defaultSettings
  } catch {
    return defaultSettings
  }
}

export function SettingsPage() {
  const [settings, setSettings] = useState<StoreSettings>(loadSettings)
  const [saved, setSaved] = useState(false)

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const sanitized: StoreSettings = {
      name: sanitize(settings.name),
      description: sanitize(settings.description),
      email: sanitize(settings.email),
      phone: sanitize(settings.phone),
      whatsapp: sanitize(settings.whatsapp),
      address: sanitize(settings.address),
      primaryColor: settings.primaryColor,
      template: settings.template,
      logo: settings.logo,
      instagram: sanitize(settings.instagram),
      facebook: sanitize(settings.facebook),
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(sanitized))
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  function update(field: keyof StoreSettings, value: string) {
    setSettings((prev) => ({ ...prev, [field]: value }))
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6 max-w-2xl">
      {/* Store Info */}
      <Card>
        <div className="flex items-center gap-2 mb-4">
          <Store size={18} className="text-primary" />
          <CardTitle>Dados da Loja</CardTitle>
        </div>
        <CardContent className="space-y-4">
          <Input
            label="Nome da Loja"
            value={settings.name}
            onChange={(e) => update('name', e.target.value)}
            placeholder="Minha Loja"
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Descricao</label>
            <textarea
              value={settings.description}
              onChange={(e) => update('description', e.target.value)}
              rows={3}
              className="w-full rounded-lg border border-border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
              placeholder="Descricao da sua loja..."
            />
          </div>
          <Input
            label="Endereco"
            value={settings.address}
            onChange={(e) => update('address', e.target.value)}
            placeholder="Rua, numero, cidade..."
          />
        </CardContent>
      </Card>

      {/* Contact */}
      <Card>
        <div className="flex items-center gap-2 mb-4">
          <Mail size={18} className="text-primary" />
          <CardTitle>Contato</CardTitle>
        </div>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Email"
              type="email"
              value={settings.email}
              onChange={(e) => update('email', e.target.value)}
              placeholder="contato@loja.com"
            />
            <Input
              label="Telefone"
              value={settings.phone}
              onChange={(e) => update('phone', e.target.value)}
              placeholder="(11) 99999-9999"
            />
          </div>
          <Input
            label="WhatsApp"
            value={settings.whatsapp}
            onChange={(e) => update('whatsapp', e.target.value)}
            placeholder="5511999999999"
          />
        </CardContent>
      </Card>

      {/* Social */}
      <Card>
        <div className="flex items-center gap-2 mb-4">
          <Globe size={18} className="text-primary" />
          <CardTitle>Redes Sociais</CardTitle>
        </div>
        <CardContent className="space-y-4">
          <Input
            label="Instagram"
            value={settings.instagram}
            onChange={(e) => update('instagram', e.target.value)}
            placeholder="@minhaloja"
          />
          <Input
            label="Facebook"
            value={settings.facebook}
            onChange={(e) => update('facebook', e.target.value)}
            placeholder="https://facebook.com/minhaloja"
          />
        </CardContent>
      </Card>

      {/* Appearance */}
      <Card>
        <div className="flex items-center gap-2 mb-4">
          <Palette size={18} className="text-primary" />
          <CardTitle>Aparencia</CardTitle>
        </div>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-1">
              <label className="block text-sm font-medium text-gray-700">Cor Principal</label>
              <div className="flex items-center gap-3">
                <input
                  type="color"
                  value={settings.primaryColor}
                  onChange={(e) => update('primaryColor', e.target.value)}
                  className="h-10 w-16 rounded border border-border cursor-pointer"
                />
                <span className="text-sm text-muted font-mono">{settings.primaryColor}</span>
              </div>
            </div>
            <Select
              label="Template"
              value={settings.template}
              onChange={(e) => update('template', e.target.value)}
              options={[
                { value: 'modern', label: 'Moderno' },
                { value: 'classic', label: 'Classico' },
                { value: 'minimal', label: 'Minimalista' },
              ]}
            />
          </div>
          <Input
            label="URL do Logo"
            value={settings.logo}
            onChange={(e) => update('logo', e.target.value)}
            placeholder="https://..."
          />
          {settings.logo && (
            <div className="mt-2">
              <img
                src={settings.logo}
                alt="Logo preview"
                className="h-16 object-contain rounded border border-border p-1"
                onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Save */}
      <div className="flex items-center gap-3">
        <Button type="submit">
          <Save size={16} /> Salvar Configuracoes
        </Button>
        {saved && (
          <span className="text-sm text-green-600 font-medium">Salvo com sucesso!</span>
        )}
      </div>
    </form>
  )
}
