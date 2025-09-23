<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-8 px-4">
    <div class="max-w-2xl mx-auto">
      <!-- Header -->
      <div class="text-center mb-8">
        <h1 class="text-4xl md:text-5xl font-bold text-gray-800 mb-3">
          🔗 URL Shortener
        </h1>
        <p class="text-lg text-gray-600">
          Shorten your long URLs quickly and easily
        </p>
      </div>

      <!-- Main Card -->
      <div class="bg-white rounded-2xl shadow-xl p-6 md:p-8 mb-6 transition-all duration-300 hover:shadow-2xl">
        <!-- Input Section -->
        <div class="mb-6">
          <label for="url-input" class="block text-sm font-medium text-gray-700 mb-2">
            Enter your long URL
          </label>
          <div class="flex flex-col sm:flex-row gap-3">
            <input
              id="url-input"
              v-model="longUrl"
              type="url"
              placeholder="https://example.com/very-long-url-path"
              class="flex-1 px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200"
              :class="{ 'border-red-300': error }"
              @keypress.enter="shortenUrl"
            />
            <button
              @click="shortenUrl"
              :disabled="loading || !longUrl"
              class="px-6 py-3 bg-primary-500 text-white font-medium rounded-lg hover:bg-primary-600 active:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 flex items-center justify-center gap-2 min-w-[120px]"
            >
              <span v-if="loading" class="animate-spin">⟳</span>
              {{ loading ? 'Shortening...' : 'Shorten' }}
            </button>
          </div>
          
          <!-- Error Message -->
          <div v-if="error" class="mt-2 text-red-600 text-sm flex items-center gap-1">
            <span>⚠</span> {{ error }}
          </div>
        </div>

        <!-- Result Section -->
        <div v-if="shortUrl" class="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg animate-fade-in">
          <p class="text-green-800 font-medium mb-2">✅ Your shortened URL:</p>
          <div class="flex flex-col sm:flex-row gap-3 items-center">
            <input
              :value="shortUrl"
              readonly
              class="flex-1 px-3 py-2 bg-white border border-green-300 rounded text-green-700 font-mono text-sm cursor-pointer"
              @click="copyToClipboard"
            />
            <button
              @click="copyToClipboard"
              class="px-4 py-2 bg-green-500 text-white text-sm rounded hover:bg-green-600 transition-colors duration-200 flex items-center gap-2"
            >
              <span>{{ copied ? 'Copied!' : 'Copy' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Features Grid -->
      <div class="grid md:grid-cols-3 gap-4 mb-8">
        <div class="bg-white/80 backdrop-blur-sm rounded-xl p-4 text-center">
          <div class="text-2xl mb-2">⚡</div>
          <h3 class="font-semibold text-gray-800 mb-1">Fast</h3>
          <p class="text-sm text-gray-600">Instant URL shortening</p>
        </div>
        <div class="bg-white/80 backdrop-blur-sm rounded-xl p-4 text-center">
          <div class="text-2xl mb-2">🔒</div>
          <h3 class="font-semibold text-gray-800 mb-1">Secure</h3>
          <p class="text-sm text-gray-600">No tracking or analytics</p>
        </div>
        <div class="bg-white/80 backdrop-blur-sm rounded-xl p-4 text-center">
          <div class="text-2xl mb-2">📱</div>
          <h3 class="font-semibold text-gray-800 mb-1">Responsive</h3>
          <p class="text-sm text-gray-600">Works on all devices</p>
        </div>
      </div>

      <!-- Recent URLs -->
      <div v-if="recentUrls.length > 0" class="bg-white/80 backdrop-blur-sm rounded-xl p-6">
        <h3 class="font-semibold text-gray-800 mb-4 flex items-center gap-2">
          <span>📚</span> Recently Shortened
        </h3>
        <div class="space-y-3 max-h-64 overflow-y-auto">
          <div
            v-for="(item, index) in recentUrls"
            :key="index"
            class="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
          >
            <div class="flex-1 min-w-0">
              <p class="text-sm text-gray-600 truncate">{{ item.original }}</p>
              <a :href="item.short" target="_blank" class="text-primary-600 hover:text-primary-700 text-sm font-mono">
                {{ item.short }}
              </a>
            </div>
            <button
              @click="copyRecentUrl(item.short)"
              class="ml-2 px-3 py-1 text-xs bg-gray-200 hover:bg-gray-300 rounded transition-colors"
            >
              Copy
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const longUrl = ref('')
const shortUrl = ref('')
const error = ref('')
const loading = ref(false)
const copied = ref(false)
const recentUrls = ref([])

const API_BASE = 'http://localhost:8080'

// Validate URL format
const isValidUrl = (string) => {
  try {
    new URL(string)
    return true
  } catch (_) {
    return false
  }
}

const shortenUrl = async () => {
  if (!longUrl.value.trim()) {
    error.value = 'Please enter a URL'
    return
  }

  if (!isValidUrl(longUrl.value)) {
    error.value = 'Please enter a valid URL (include http:// or https://)'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`${API_BASE}/shorten`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url: longUrl.value }),
    })

    if (!response.ok) {
      throw new Error('Failed to shorten URL')
    }

    const data = await response.json()
    shortUrl.value = data.short_url
    
    // Add to recent URLs
    recentUrls.value.unshift({
      original: longUrl.value,
      short: data.short_url
    })
    
    // Keep only last 5 URLs
    if (recentUrls.value.length > 5) {
      recentUrls.value = recentUrls.value.slice(0, 5)
    }
    
    // Save to localStorage
    localStorage.setItem('recentUrls', JSON.stringify(recentUrls.value))
    
  } catch (err) {
    error.value = 'Failed to shorten URL. Please check if the server is running.'
    console.error('Error:', err)
  } finally {
    loading.value = false
  }
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(shortUrl.value)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    // Fallback for older browsers
    const textArea = document.createElement('textarea')
    textArea.value = shortUrl.value
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  }
}

const copyRecentUrl = (url) => {
  navigator.clipboard.writeText(url)
  // You could add a toast notification here
}

// Load recent URLs from localStorage on component mount
onMounted(() => {
  const saved = localStorage.getItem('recentUrls')
  if (saved) {
    recentUrls.value = JSON.parse(saved)
  }
})
</script>

<style>
.animate-fade-in {
  animation: fadeIn 0.5s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>