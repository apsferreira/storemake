import type { TemplateTheme } from '../types';

export interface ThemeConfig {
  name: string;
  label: string;
  primaryColor: string;
  secondaryColor: string;
  accentColor: string;
  bgColor: string;
  cardBg: string;
  textColor: string;
  fontFamily: string;
  borderRadius: string;
  buttonStyle: string;
  headerBg: string;
}

export const themes: Record<TemplateTheme, ThemeConfig> = {
  moda: {
    name: 'moda',
    label: 'Moda',
    primaryColor: '#1a1a1a',
    secondaryColor: '#f5f5f5',
    accentColor: '#c9a96e',
    bgColor: '#ffffff',
    cardBg: '#ffffff',
    textColor: '#1a1a1a',
    fontFamily: "'Inter', sans-serif",
    borderRadius: '0px',
    buttonStyle: 'bg-black text-white hover:bg-gray-800',
    headerBg: 'bg-white border-b border-gray-200',
  },
  'semi-joias': {
    name: 'semi-joias',
    label: 'Semi-joias',
    primaryColor: '#8b6f47',
    secondaryColor: '#faf7f2',
    accentColor: '#d4a853',
    bgColor: '#faf7f2',
    cardBg: '#ffffff',
    textColor: '#3d2b1f',
    fontFamily: "'Playfair Display', serif",
    borderRadius: '8px',
    buttonStyle: 'bg-amber-700 text-white hover:bg-amber-800',
    headerBg: 'bg-amber-50 border-b border-amber-200',
  },
  festas: {
    name: 'festas',
    label: 'Festas',
    primaryColor: '#e91e8c',
    secondaryColor: '#fff0f7',
    accentColor: '#ff6bb5',
    bgColor: '#fff0f7',
    cardBg: '#ffffff',
    textColor: '#2d1b33',
    fontFamily: "'Poppins', sans-serif",
    borderRadius: '16px',
    buttonStyle: 'bg-pink-600 text-white hover:bg-pink-700',
    headerBg: 'bg-pink-50 border-b border-pink-200',
  },
  artesanato: {
    name: 'artesanato',
    label: 'Artesanato',
    primaryColor: '#2d6a4f',
    secondaryColor: '#f0f7f4',
    accentColor: '#52b788',
    bgColor: '#f0f7f4',
    cardBg: '#ffffff',
    textColor: '#1b4332',
    fontFamily: "'Nunito', sans-serif",
    borderRadius: '12px',
    buttonStyle: 'bg-emerald-700 text-white hover:bg-emerald-800',
    headerBg: 'bg-emerald-50 border-b border-emerald-200',
  },
};

export function getTheme(name: TemplateTheme): ThemeConfig {
  return themes[name] || themes.moda;
}
