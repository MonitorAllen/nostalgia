import { describe, expect, test } from 'bun:test'
import {
  ADMIN_ARTICLES_PATH,
  ADMIN_BASE_PATH,
  ADMIN_LOGIN_PATH,
  buildAdminLoginRedirect
} from './adminRoutes'

describe('adminRoutes', () => {
  test('uses backend as the canonical admin base path', () => {
    expect(ADMIN_BASE_PATH).toBe('/backend')
    expect(ADMIN_LOGIN_PATH).toBe('/backend/login')
    expect(ADMIN_ARTICLES_PATH).toBe('/backend/articles')
  })

  test('does not add a redirect query when already on the login page', () => {
    expect(buildAdminLoginRedirect('/backend/login')).toBe('/backend/login')
    expect(buildAdminLoginRedirect('/backend/login?redirect=%2Fbackend%2Farticles')).toBe(
      '/backend/login'
    )
  })

  test('preserves protected destination as an encoded redirect query', () => {
    expect(buildAdminLoginRedirect('/backend/articles?page=2#top')).toBe(
      '/backend/login?redirect=%2Fbackend%2Farticles%3Fpage%3D2%23top'
    )
  })
})
