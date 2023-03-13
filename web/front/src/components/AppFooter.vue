<template>
  <CFooter>
    <div>
      <a
        href="https://github.com/ipfs-force-community/brightbird"
        target="_blank"
      >
        BrightBird
      </a>
      <span class="ms-1"
        >&copy; {{ new Date().getFullYear() }}
        Venus QA Team.
      </span>
    </div>
    <div class="ms-auto">
      <span class="me-1" target="_blank">Powered by</span>
      <a href="https://coreui.io/vue">IPFS Community</a>
    </div>
  </CFooter>
</template>

<script>
export default {
  name: 'AppFooter',
}
</script>
