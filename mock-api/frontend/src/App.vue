<script setup>
import { ref, onMounted } from 'vue';

const mocks = ref([]);
const loading = ref(false);
const error = ref(null);

const newMock = ref({
  path: '',
  method: 'GET',
  responseStatusCode: 200,
  responseBody: '{}',
  contentType: 'application/json',
});

// Función para cargar los mocks
const fetchMocks = async () => {
  loading.value = true;
  error.value = null;
  try {
    const response = await fetch('http://localhost:3000/configure-mock');
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    mocks.value = await response.json();
  } catch (e) {
    error.value = 'Error al cargar los mocks: ' + e.message;
    console.error('Error fetching mocks:', e);
  } finally {
    loading.value = false;
  }
};

// Función para agregar un mock
const addMock = async () => {
  try {
    // Asegúrate de que responseBody sea un objeto o array, no un string
    const bodyParsed = JSON.parse(newMock.value.responseBody);

    const response = await fetch('http://localhost:3000/configure-mock', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        path: newMock.value.path,
        method: newMock.value.method.toUpperCase(),
        responseStatusCode: newMock.value.responseStatusCode,
        responseBody: bodyParsed,
        contentType: newMock.value.contentType,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(`HTTP error! status: ${response.status} - ${errorData.error || response.statusText}`);
    }

    const result = await response.json();
    showAlert(result.message, 'success');
    // Limpiar formulario y recargar mocks
    newMock.value = {
      path: '',
      method: 'GET',
      responseStatusCode: 200,
      responseBody: '{}',
      contentType: 'application/json',
    };
    fetchMocks();
  } catch (e) {
    showAlert('Error al agregar mock: ' + e.message, 'danger');
    console.error('Error adding mock:', e);
  }
};

// Función para eliminar un mock
const deleteMock = async (id) => {
    if (!confirm(`¿Estás seguro de que quieres eliminar el mock con ID: ${id}?`)) {
        return;
    }
    try {
        const response = await fetch(`http://localhost:3000/configure-mock/${id}`, {
            method: 'DELETE',
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(`HTTP error! status: ${response.status} - ${errorData.error || response.statusText}`);
        }

        const result = await response.json();
        showAlert(result.message, 'success');
        fetchMocks(); // Recargar la lista después de eliminar
    } catch (e) {
        showAlert('Error al eliminar mock: ' + e.message, 'danger');
        console.error('Error deleting mock:', e);
    }
};

// Función para mostrar alertas de Bootstrap
const showAlert = (message, type) => {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    document.querySelector('.alert-container').appendChild(alertDiv);
    
    // Auto-remove alert after 5 seconds
    setTimeout(() => {
        alertDiv.remove();
    }, 5000);
};

// Función para obtener el color de badge según el método HTTP
const getMethodBadgeClass = (method) => {
  const methodColors = {
    'GET': 'bg-success',
    'POST': 'bg-primary',
    'PUT': 'bg-warning',
    'DELETE': 'bg-danger',
    'PATCH': 'bg-info'
  };
  return methodColors[method] || 'bg-secondary';
};

// Función para obtener el color de badge según el status code
const getStatusBadgeClass = (status) => {
  if (status >= 200 && status < 300) return 'bg-success';
  if (status >= 300 && status < 400) return 'bg-info';
  if (status >= 400 && status < 500) return 'bg-warning';
  if (status >= 500) return 'bg-danger';
  return 'bg-secondary';
};

// Cargar mocks al montar el componente
onMounted(fetchMocks);
</script>

<template>
  <div id="app">
    <div class="container py-4">
      <!-- Header -->
      <div class="row mb-4">
        <div class="col-12 text-center">
          <h1 class="display-5 fw-bold text-primary mb-3">
            <i class="bi bi-gear-fill me-2"></i>
            Gestor de Mocks API
          </h1>
          <p class="lead text-muted">Configura y gestiona tus endpoints mock de forma sencilla</p>
        </div>
      </div>

      <!-- Alert Container -->
      <div class="alert-container mb-4"></div>

      <!-- Formulario para nuevo mock -->
      <div class="row justify-content-center mb-4">
        <div class="col-12 col-lg-8">
          <div class="card shadow">
            <div class="card-header bg-primary text-white">
              <h4 class="card-title mb-0">
                <i class="bi bi-plus-circle me-2"></i>
                Configurar Nuevo Mock
              </h4>
            </div>
            <div class="card-body">
              <form @submit.prevent="addMock">
                <div class="row g-3">
                  <!-- Path y Método en la misma fila -->
                  <div class="col-sm-8">
                    <label for="path" class="form-label fw-semibold">Path del Endpoint</label>
                    <div class="input-group">
                      <span class="input-group-text">
                        <i class="bi bi-link-45deg"></i>
                      </span>
                      <input 
                        id="path"
                        v-model="newMock.path" 
                        type="text"
                        class="form-control" 
                        placeholder="/api/users" 
                        required 
                      />
                    </div>
                  </div>
                  
                  <div class="col-sm-4">
                    <label for="method" class="form-label fw-semibold">Método HTTP</label>
                    <select 
                      id="method"
                      v-model="newMock.method" 
                      class="form-select"
                      required
                    >
                      <option value="GET">GET</option>
                      <option value="POST">POST</option>
                      <option value="PUT">PUT</option>
                      <option value="DELETE">DELETE</option>
                      <option value="PATCH">PATCH</option>
                    </select>
                  </div>
                </div>

                <div class="row g-3 mt-2">
                  <!-- Status Code y Content Type -->
                  <div class="col-sm-6">
                    <label for="statusCode" class="form-label fw-semibold">Código de Estado</label>
                    <div class="input-group">
                      <span class="input-group-text">
                        <i class="bi bi-hash"></i>
                      </span>
                      <input 
                        id="statusCode"
                        v-model.number="newMock.responseStatusCode" 
                        type="number" 
                        class="form-control" 
                        placeholder="200" 
                        min="100" 
                        max="599"
                        required 
                      />
                    </div>
                  </div>
                  
                  <div class="col-sm-6">
                    <label for="contentType" class="form-label fw-semibold">Content Type</label>
                    <div class="input-group">
                      <span class="input-group-text">
                        <i class="bi bi-file-earmark-code"></i>
                      </span>
                      <input 
                        id="contentType"
                        v-model="newMock.contentType" 
                        type="text"
                        class="form-control" 
                        placeholder="application/json" 
                        required 
                      />
                    </div>
                  </div>
                </div>

                <div class="row g-3 mt-2">
                  <!-- Response Body -->
                  <div class="col-12">
                    <label for="responseBody" class="form-label fw-semibold">Response Body (JSON)</label>
                    <textarea 
                      id="responseBody"
                      v-model="newMock.responseBody" 
                      class="form-control font-monospace" 
                      rows="5"
                      placeholder='{"message": "Hello World", "data": []}'
                      required
                    ></textarea>
                  </div>
                </div>

                <div class="row g-3 mt-3">
                  <div class="col-12 d-grid">
                    <button type="submit" class="btn btn-primary">
                      <i class="bi bi-plus-circle me-2"></i>
                      Agregar Mock
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>

      <!-- Lista de mocks -->
      <div class="row justify-content-center">
        <div class="col-12 col-lg-8">
          <div class="card shadow">
            <div class="card-header bg-dark text-white d-flex justify-content-between align-items-center">
              <h4 class="card-title mb-0">
                <i class="bi bi-list-ul me-2"></i>
                Mocks Configurados
              </h4>
              <button @click="fetchMocks" class="btn btn-outline-light btn-sm">
                <i class="bi bi-arrow-clockwise me-1"></i>
                Actualizar
              </button>
            </div>
            <div class="card-body">
              <!-- Loading -->
              <div v-if="loading" class="text-center py-5">
                <div class="spinner-border text-primary" role="status">
                  <span class="visually-hidden">Cargando...</span>
                </div>
                <p class="mt-3 text-muted">Cargando mocks...</p>
              </div>

              <!-- Error -->
              <div v-else-if="error" class="alert alert-danger" role="alert">
                <i class="bi bi-exclamation-triangle me-2"></i>
                {{ error }}
              </div>

              <!-- Empty state -->
              <div v-else-if="mocks.length === 0" class="text-center py-5">
                <i class="bi bi-inbox display-1 text-muted"></i>
                <h4 class="mt-3 text-muted">No hay mocks configurados</h4>
                <p class="text-muted">Agrega tu primer mock usando el formulario de arriba</p>
              </div>

              <!-- Mocks list -->
              <div v-else class="row g-3">
                <div v-for="mock in mocks" :key="mock.id" class="col-12 col-sm-6">
                  <div class="card h-100 border-start border-primary border-4">
                    <div class="card-body d-flex flex-column">
                      <div class="d-flex justify-content-between align-items-start mb-3">
                        <div class="flex-grow-1">
                          <h6 class="card-title mb-2">
                            <span :class="`badge ${getMethodBadgeClass(mock.method)} me-2`">
                              {{ mock.method }}
                            </span>
                          </h6>
                          <div class="mb-2">
                            <code class="text-dark small d-block text-break">{{ mock.path }}</code>
                          </div>
                          <div class="mb-2">
                            <small class="text-muted">Status:</small>
                            <span :class="`badge ${getStatusBadgeClass(mock.responseStatusCode)} ms-1`">
                              {{ mock.responseStatusCode }}
                            </span>
                          </div>
                        </div>
                        <button 
                          @click="deleteMock(mock.id)"
                          class="btn btn-outline-danger btn-sm ms-2"
                          title="Eliminar mock"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                      
                      <div class="mb-2">
                        <small class="text-muted d-block mb-1">ID:</small>
                        <code class="text-info small text-break">{{ mock.id }}</code>
                      </div>

                      <div class="mb-3">
                        <small class="text-muted d-block mb-1">Content Type:</small>
                        <span class="badge bg-light text-dark small">{{ mock.contentType }}</span>
                      </div>

                      <div class="flex-grow-1">
                        <small class="text-muted d-block mb-2">Response Body:</small>
                        <pre class="bg-light p-2 rounded border overflow-auto" style="max-height: 120px; font-size: 0.75rem;"><code>{{ typeof mock.responseBody === 'string' ? mock.responseBody : JSON.stringify(mock.responseBody, null, 2) }}</code></pre>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Estilos mínimos - usando principalmente Bootstrap */
.card {
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
}

.card:hover {
  transform: translateY(-2px);
}

/* Iconos de Bootstrap */
@import url('https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.0/font/bootstrap-icons.css');
</style>