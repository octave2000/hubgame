import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from 'react'

type Game = {
  id: string
  name: string
  studio: string
  category: 'Strategy' | 'Puzzle' | 'Party' | 'Arcade' | 'Board' | 'Adventure'
  modes: Array<'Solo' | 'Online Multiplayer' | 'Offline Multiplayer' | 'Bot Match'>
  vibe: string
  cover: string
  rating: number
  players: string
  installed: boolean
  tags: string[]
}

type EntityRecord = {
  id: string
  data: Partial<Game>
}

const seedCatalog: Game[] = [
  {
    id: 'mod-grid',
    name: 'Modifier Grid',
    studio: 'Northline Labs',
    category: 'Strategy',
    modes: ['Solo', 'Online Multiplayer', 'Bot Match'],
    vibe: 'Sharp tactical rounds with playful modifiers.',
    cover:
      'https://images.unsplash.com/photo-1511512578047-dfb367046420?auto=format&fit=crop&w=1800&q=80',
    rating: 4.8,
    players: '42k online',
    installed: true,
    tags: ['ranked', 'modifiers', 'pvp'],
  },
  {
    id: 'rift-racers',
    name: 'Rift Racers',
    studio: 'Quiet Orbit',
    category: 'Arcade',
    modes: ['Online Multiplayer', 'Offline Multiplayer'],
    vibe: 'Fast rounds, smooth drifts, bright rush moments.',
    cover:
      'https://images.unsplash.com/photo-1579373903781-fd5c0c30c4cd?auto=format&fit=crop&w=1800&q=80',
    rating: 4.6,
    players: '18k online',
    installed: false,
    tags: ['racing', 'party', 'quick'],
  },
  {
    id: 'tiny-kingdoms',
    name: 'Tiny Kingdoms',
    studio: 'Clay Beacon',
    category: 'Strategy',
    modes: ['Solo', 'Bot Match'],
    vibe: 'Small maps with surprisingly deep decisions.',
    cover:
      'https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&w=1800&q=80',
    rating: 4.7,
    players: '11k online',
    installed: false,
    tags: ['simulation', 'deep', 'solo'],
  },
  {
    id: 'echo-tiles',
    name: 'Echo Tiles',
    studio: 'Moss Arcade',
    category: 'Puzzle',
    modes: ['Solo', 'Online Multiplayer'],
    vibe: 'Rhythm and puzzle flow for calm sessions.',
    cover:
      'https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&w=1800&q=80',
    rating: 4.5,
    players: '9k online',
    installed: true,
    tags: ['relaxing', 'co-op', 'focus'],
  },
  {
    id: 'couch-club',
    name: 'Couch Club',
    studio: 'Gather House',
    category: 'Party',
    modes: ['Offline Multiplayer', 'Online Multiplayer'],
    vibe: 'Light mini-games for local nights.',
    cover:
      'https://images.unsplash.com/photo-1511882150382-421056c89033?auto=format&fit=crop&w=1800&q=80',
    rating: 4.4,
    players: '24k online',
    installed: false,
    tags: ['friends', 'casual', 'fun'],
  },
  {
    id: 'canvas-quest',
    name: 'Canvas Quest',
    studio: 'Signal Pine',
    category: 'Adventure',
    modes: ['Solo', 'Online Multiplayer'],
    vibe: 'Warm world exploration with cooperative moments.',
    cover:
      'https://images.unsplash.com/photo-1493711662062-fa541adb3fc8?auto=format&fit=crop&w=1800&q=80',
    rating: 4.9,
    players: '31k online',
    installed: false,
    tags: ['story', 'co-op', 'exploration'],
  },
]

const categories = ['All', 'Strategy', 'Puzzle', 'Party', 'Arcade', 'Board', 'Adventure'] as const
const modes = ['All', 'Solo', 'Online Multiplayer', 'Offline Multiplayer', 'Bot Match'] as const
const sortOptions = ['Curated', 'Top Rated', 'Most Active', 'A-Z'] as const

const gatewayURL = import.meta.env.VITE_GATEWAY_URL ?? 'http://localhost:8080'
const devTenantID = import.meta.env.VITE_DEV_TENANT_ID ?? 'hubgame-dev'
const devUserID = import.meta.env.VITE_DEV_USER_ID ?? 'web-dev-user'
const devRole = import.meta.env.VITE_DEV_ROLE ?? 'developer'
const tokenStorageKey = 'hubgame.dev.token'

function App() {
  const wsRef = useRef<WebSocket | null>(null)
  const [pointer, setPointer] = useState({ x: 50, y: 50 })
  const [query, setQuery] = useState('')
  const [category, setCategory] = useState<(typeof categories)[number]>('All')
  const [mode, setMode] = useState<(typeof modes)[number]>('All')
  const [installedOnly, setInstalledOnly] = useState(false)
  const [sortBy, setSortBy] = useState<(typeof sortOptions)[number]>('Curated')
  const [showLookup, setShowLookup] = useState(false)
  const [selectedGame, setSelectedGame] = useState<Game | null>(null)
  const [games, setGames] = useState<Game[]>([])
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(true)
  const [syncing, setSyncing] = useState(false)
  const [error, setError] = useState('')

  const loadCatalog = useCallback(async (authToken: string) => {
    const response = await fetch(`${gatewayURL}/v1/entities?kind=game`, {
      headers: { Authorization: `Bearer ${authToken}` },
    })
    if (!response.ok) {
      throw new Error(`Catalog request failed (${response.status})`)
    }
    const entities = (await response.json()) as EntityRecord[]
    const catalog = entities
      .map((entity) => toGame(entity.id, entity.data))
      .filter((game): game is Game => game !== null)
    setGames(catalog)
    return catalog
  }, [])

  const syncCatalog = useCallback(
    async (authToken: string, silent = false) => {
      if (!silent) {
        setSyncing(true)
      }
      setError('')
      try {
        await loadCatalog(authToken)
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to sync catalog'
        setError(message)
      } finally {
        setSyncing(false)
      }
    },
    [loadCatalog],
  )

  const connectRealtime = useCallback(
    (authToken: string) => {
      if (wsRef.current) {
        wsRef.current.close()
      }
      const wsURL = gatewayURL.replace(/^http/, 'ws')
      const socket = new WebSocket(`${wsURL}/v1/events/stream?topic=entity.game&access_token=${encodeURIComponent(authToken)}`)
      wsRef.current = socket

      socket.onmessage = () => {
        void syncCatalog(authToken, true)
      }

      socket.onclose = () => {
        if (wsRef.current === socket) {
          wsRef.current = null
        }
      }
    },
    [syncCatalog],
  )

  const ensureToken = useCallback(async () => {
    const cached = localStorage.getItem(tokenStorageKey)
    if (cached) {
      return cached
    }

    const response = await fetch(`${gatewayURL}/v1/auth/dev-token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: devUserID, tenant_id: devTenantID, role: devRole, ttl_seconds: 86400 }),
    })
    if (!response.ok) {
      throw new Error(`Unable to issue dev token (${response.status}). Enable HUBGAME_ENABLE_DEV_AUTH on gateway.`)
    }
    const payload = (await response.json()) as { token?: string }
    if (!payload.token) {
      throw new Error('Gateway did not return a token')
    }
    localStorage.setItem(tokenStorageKey, payload.token)
    return payload.token
  }, [])

  const seedCatalogToBackend = useCallback(async () => {
    if (!token) {
      return
    }
    setSyncing(true)
    setError('')
    try {
      for (const game of seedCatalog) {
        const response = await fetch(`${gatewayURL}/v1/entities`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({ id: game.id, kind: 'game', data: game }),
        })
        if (response.ok) {
          continue
        }
        const text = await response.text()
        if (!text.includes('UNIQUE constraint failed')) {
          throw new Error(`Seed failed for ${game.name}: ${text || response.statusText}`)
        }
      }
      await syncCatalog(token)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to seed catalog'
      setError(message)
    } finally {
      setSyncing(false)
    }
  }, [syncCatalog, token])

  useEffect(() => {
    let alive = true

    const init = async () => {
      setLoading(true)
      setError('')
      try {
        const authToken = await ensureToken()
        if (!alive) {
          return
        }
        setToken(authToken)
        await syncCatalog(authToken)
        if (!alive) {
          return
        }
        connectRealtime(authToken)
      } catch (err) {
        if (!alive) {
          return
        }
        const message = err instanceof Error ? err.message : 'Failed to initialize store'
        setError(message)
      } finally {
        if (alive) {
          setLoading(false)
        }
      }
    }

    void init()

    return () => {
      alive = false
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [connectRealtime, ensureToken, syncCatalog])

  const filteredGames = useMemo(() => {
    const normalized = query.trim().toLowerCase()

    const result = games.filter((game) => {
      const matchQuery =
        normalized.length === 0 ||
        [game.name, game.studio, game.vibe, ...game.tags].join(' ').toLowerCase().includes(normalized)
      const matchCategory = category === 'All' || game.category === category
      const matchMode = mode === 'All' || game.modes.includes(mode)
      const matchInstalled = !installedOnly || game.installed
      return matchQuery && matchCategory && matchMode && matchInstalled
    })

    result.sort((a, b) => {
      if (sortBy === 'Top Rated') return b.rating - a.rating
      if (sortBy === 'Most Active') return b.players.localeCompare(a.players)
      if (sortBy === 'A-Z') return a.name.localeCompare(b.name)
      return 0
    })

    return result
  }, [category, games, installedOnly, mode, query, sortBy])

  const featured = filteredGames[0]
  const shellStyle = {
    '--mx': `${pointer.x}%`,
    '--my': `${pointer.y}%`,
  } as CSSProperties

  return (
    <div
      className="store-shell min-h-screen text-stone-800"
      style={shellStyle}
      onMouseMove={(event) => {
        const rect = event.currentTarget.getBoundingClientRect()
        const x = ((event.clientX - rect.left) / rect.width) * 100
        const y = ((event.clientY - rect.top) / rect.height) * 100
        setPointer({ x, y })
      }}
      onMouseLeave={() => setPointer({ x: 50, y: 50 })}
    >
      <div className="store-glow store-glow-a" />
      <div className="store-glow store-glow-b" />
      <div className="store-grid" />
      <div className="store-grid-reactive" />
      <div className="store-grid-fine" />

      <main className="relative mx-auto max-w-[1300px] px-4 pb-14 pt-8 sm:px-6 lg:px-8">
        <header className="reveal-up mb-7 rounded-[30px] border border-[#d9c9b2] bg-[#f2e7d8]/85 p-5 shadow-[0_16px_40px_rgba(82,57,27,0.08)] backdrop-blur-sm sm:p-7">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <p className="text-xs uppercase tracking-[0.25em] text-stone-600">HubGame Store</p>
              <h1 className="font-display mt-2 text-3xl leading-tight text-[#3f2f21] sm:text-5xl">Find your next cozy match</h1>
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs">
              <span className="rounded-full border border-[#b99f80] bg-[#e8d7c3] px-3 py-1 text-[#5d4124]">
                {token ? 'Gateway Connected' : 'Disconnected'}
              </span>
              <button
                onClick={() => {
                  if (!token) return
                  void syncCatalog(token)
                }}
                className="rounded-full border border-[#b89c78] bg-[#ecd8bf] px-4 py-2 text-sm text-[#5d4124] transition hover:-translate-y-0.5 hover:bg-[#e7cfb1]"
              >
                {syncing ? 'Syncing...' : 'Sync'}
              </button>
              <button
                onClick={() => setShowLookup((value) => !value)}
                className="rounded-full border border-[#b89c78] bg-[#ecd8bf] px-4 py-2 text-sm text-[#5d4124] transition hover:-translate-y-0.5 hover:bg-[#e7cfb1]"
              >
                {showLookup ? 'Hide Lookup' : 'Open Lookup'}
              </button>
            </div>
          </div>

          {error ? <p className="mt-4 rounded-xl border border-[#c59f6c] bg-[#f4e6d2] px-3 py-2 text-sm text-[#6c431a]">{error}</p> : null}

          {featured ? (
            <section className="mt-5 overflow-hidden rounded-3xl border border-[#c9b193] bg-[#24190f]">
              <img src={featured.cover} alt={featured.name} className="h-[48vh] min-h-[340px] w-full object-cover opacity-75" />
              <div className="-mt-40 flex flex-col gap-3 p-5 text-[#f4ecde] sm:p-8">
                <p className="text-xs uppercase tracking-[0.22em] text-[#d6c2a6]">Featured Today</p>
                <h2 className="font-display text-4xl sm:text-5xl">{featured.name}</h2>
                <p className="max-w-2xl text-sm text-[#e8dac6]">{featured.vibe}</p>
                <div className="mt-2 flex flex-wrap items-center gap-3 text-xs text-[#dec7a7]">
                  <span>★ {featured.rating.toFixed(1)}</span>
                  <span>{featured.players}</span>
                  <span>{featured.studio}</span>
                </div>
              </div>
            </section>
          ) : (
            <section className="mt-5 rounded-3xl border border-[#c9b193] bg-[#f0dfc9] p-6 text-[#63462a]">
              <h2 className="font-display text-3xl">Catalog is empty</h2>
              <p className="mt-2 text-sm">Seed a starter catalog into backend and this store will become live immediately.</p>
              <button
                onClick={() => void seedCatalogToBackend()}
                className="mt-4 rounded-xl bg-[#6d4c2e] px-4 py-2 text-sm text-[#fff4e7] transition hover:bg-[#5f4127]"
              >
                Seed Starter Catalog
              </button>
            </section>
          )}

          <div
            className={`grid transition-all duration-500 ${
              showLookup ? 'mt-5 max-h-[380px] grid-rows-[1fr] opacity-100' : 'max-h-0 grid-rows-[0fr] opacity-0'
            }`}
          >
            <div className="overflow-hidden rounded-2xl border border-[#cfb79b] bg-[#efe2d0] p-4">
              <div className="grid gap-3 md:grid-cols-[2fr_1fr_1fr]">
                <input
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Search title, vibe, tag or studio"
                  className="rounded-xl border border-[#ceb08f] bg-[#f7efe3] px-4 py-3 text-sm outline-none focus:border-[#9c774c]"
                />
                <select
                  value={mode}
                  onChange={(event) => setMode(event.target.value as (typeof modes)[number])}
                  className="rounded-xl border border-[#ceb08f] bg-[#f7efe3] px-3 py-3 text-sm"
                >
                  {modes.map((item) => (
                    <option key={item}>{item}</option>
                  ))}
                </select>
                <select
                  value={sortBy}
                  onChange={(event) => setSortBy(event.target.value as (typeof sortOptions)[number])}
                  className="rounded-xl border border-[#ceb08f] bg-[#f7efe3] px-3 py-3 text-sm"
                >
                  {sortOptions.map((item) => (
                    <option key={item}>{item}</option>
                  ))}
                </select>
              </div>
              <label className="mt-3 inline-flex items-center gap-2 text-sm text-[#6a4c2d]">
                <input type="checkbox" checked={installedOnly} onChange={() => setInstalledOnly((value) => !value)} />
                Installed only
              </label>
            </div>
          </div>
        </header>

        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap gap-2">
            {categories.map((item) => {
              const active = category === item
              return (
                <button
                  key={item}
                  onClick={() => setCategory(item)}
                  className={`rounded-full border px-4 py-2 text-sm transition ${
                    active
                      ? 'border-[#715133] bg-[#7f5e3d] text-[#fff6ea] shadow-[0_10px_25px_rgba(66,44,21,0.2)]'
                      : 'border-[#cfb79b] bg-[#efdfca] text-[#6e4f31] hover:bg-[#ead6bd]'
                  }`}
                >
                  {item}
                </button>
              )
            })}
          </div>
          <p className="text-sm text-[#68492c]">
            {loading ? 'Loading catalog...' : `${filteredGames.length} games`}
          </p>
        </div>

        <section className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          {filteredGames.map((game, index) => (
            <button
              key={game.id}
              type="button"
              onClick={() => setSelectedGame(game)}
              className="reveal-up group overflow-hidden rounded-[26px] border border-[#ccb18f] bg-[#f1e2cf]/95 text-left shadow-[0_18px_30px_rgba(66,44,21,0.09)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_24px_36px_rgba(66,44,21,0.16)]"
              style={{ animationDelay: `${index * 70}ms` }}
            >
              <div className="relative overflow-hidden">
                <img
                  src={game.cover}
                  alt={game.name}
                  className="h-72 w-full object-cover transition duration-500 group-hover:scale-[1.04]"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/15 to-transparent" />
                <div className="absolute bottom-4 left-4 right-4 text-[#f6ecdf]">
                  <h3 className="font-display text-3xl leading-none">{game.name}</h3>
                  <p className="mt-1 text-xs uppercase tracking-[0.16em] text-[#e2ceaf]">{game.studio}</p>
                </div>
              </div>
              <div className="flex items-center justify-between px-4 py-3 text-sm text-[#654729]">
                <div className="flex items-center gap-3">
                  <span>★ {game.rating.toFixed(1)}</span>
                  <span>{game.players}</span>
                </div>
                <span className="rounded-full border border-[#be9f79] bg-[#ead7bf] px-3 py-1 text-xs">Play</span>
              </div>
            </button>
          ))}
        </section>
      </main>

      {selectedGame ? (
        <aside className="fixed inset-0 z-40 flex items-end bg-black/45 p-0 sm:items-center sm:justify-center sm:p-6" onClick={() => setSelectedGame(null)}>
          <div
            className="reveal-up w-full max-w-3xl overflow-hidden rounded-t-[28px] border border-[#ccb18f] bg-[#f2e5d2] shadow-[0_30px_60px_rgba(32,18,8,0.35)] sm:rounded-[30px]"
            onClick={(event) => event.stopPropagation()}
          >
            <img src={selectedGame.cover} alt={selectedGame.name} className="h-72 w-full object-cover" />
            <div className="space-y-4 p-5 sm:p-7">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <h3 className="font-display text-4xl text-[#3f2f21]">{selectedGame.name}</h3>
                  <p className="text-sm uppercase tracking-[0.16em] text-[#7b5a3a]">{selectedGame.studio}</p>
                </div>
                <button
                  onClick={() => setSelectedGame(null)}
                  className="rounded-full border border-[#b99b78] bg-[#e9d5bc] px-4 py-2 text-sm text-[#5f4327]"
                >
                  Close
                </button>
              </div>

              <p className="text-[#5d4024]">{selectedGame.vibe}</p>

              <div className="flex flex-wrap gap-2">
                {selectedGame.modes.map((item) => (
                  <span key={item} className="rounded-full border border-[#c9ab86] bg-[#ecdbc6] px-3 py-1 text-xs text-[#644628]">
                    {item}
                  </span>
                ))}
              </div>

              <div className="flex flex-wrap items-center gap-3 text-sm text-[#66492c]">
                <span>★ {selectedGame.rating.toFixed(1)}</span>
                <span>{selectedGame.players}</span>
                <span>{selectedGame.category}</span>
              </div>

              <button className="rounded-xl bg-[#6d4c2e] px-5 py-3 text-sm font-semibold text-[#fff4e7] transition hover:bg-[#5f4127]">
                {selectedGame.installed ? 'Open Game' : 'Install Game'}
              </button>
            </div>
          </div>
        </aside>
      ) : null}
    </div>
  )
}

function toGame(id: string, data: Partial<Game>): Game | null {
  if (!data.name || !data.cover || !data.studio || !data.category || !data.modes || !data.vibe) {
    return null
  }
  return {
    id,
    name: data.name,
    cover: data.cover,
    studio: data.studio,
    category: data.category,
    modes: data.modes,
    vibe: data.vibe,
    rating: data.rating ?? 4.2,
    players: data.players ?? 'new',
    installed: data.installed ?? false,
    tags: data.tags ?? [],
  }
}

export default App
