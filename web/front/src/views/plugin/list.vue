<template>
  <CContainer fluid>
    <CRow>
      <CCol :xs="9">
        <CCard class="mb-4">
          <CCardHeader>
            <strong>Vue Table</strong> <small>Basic example</small>
          </CCardHeader>
          <CCardBody>
            <p class="text-medium-emphasis small">
              Using the most basic table CoreUI, here&#39;s how
              <code>&lt;CTable&gt;</code>-based tables look in CoreUI.
            </p>
            <DocsExample href="components/table.html">
              <CTable>
                <CTableHead>
                  <CTableRow>
                    <CTableHeaderCell scope="col">#</CTableHeaderCell>
                    <CTableHeaderCell scope="col">Class</CTableHeaderCell>
                    <CTableHeaderCell scope="col">Heading</CTableHeaderCell>
                    <CTableHeaderCell scope="col">Heading</CTableHeaderCell>
                  </CTableRow>
                </CTableHead>
                <CTableBody>
                  <CTableRow>
                    <CTableHeaderCell scope="row">1</CTableHeaderCell>
                    <CTableDataCell>Mark</CTableDataCell>
                    <CTableDataCell>Otto</CTableDataCell>
                    <CTableDataCell>@mdo</CTableDataCell>
                  </CTableRow>
                  <CTableRow>
                    <CTableHeaderCell scope="row">2</CTableHeaderCell>
                    <CTableDataCell>Jacob</CTableDataCell>
                    <CTableDataCell>Thornton</CTableDataCell>
                    <CTableDataCell>@fat</CTableDataCell>
                  </CTableRow>
                  <CTableRow>
                    <CTableHeaderCell scope="row">3</CTableHeaderCell>
                    <CTableDataCell colspan="2">Larry the Bird</CTableDataCell>
                    <CTableDataCell>@twitter</CTableDataCell>
                  </CTableRow>
                </CTableBody>
              </CTable>
            </DocsExample>
          </CCardBody>
        </CCard>
      </CCol>
      <CCol :xs="3">
        <CForm class="row g-3">
          <CCol :md="6">
            <CFormLabel for="inputEmail4">Email</CFormLabel>
            <CFormInput id="inputEmail4" type="email" />
          </CCol>
          <CCol :md="6">
            <CFormLabel for="inputPassword4">Password</CFormLabel>
            <CFormInput id="inputPassword4" type="password" />
          </CCol>
          <CCol :xs="12">
            <CFormLabel for="inputAddress">Address</CFormLabel>
            <CFormInput id="inputAddress" placeholder="1234 Main St" />
          </CCol>
          <CCol :xs="12">
            <CFormLabel for="inputAddress2">Address 2</CFormLabel>
            <CFormInput
              id="inputAddress2"
              placeholder="Apartment, studio, or floor"
            />
          </CCol>
          <CCol :md="6">
            <CFormLabel for="inputCity">City</CFormLabel>
            <CFormInput id="inputCity" />
          </CCol>
          <CCol :md="4">
            <CFormLabel for="inputState">State</CFormLabel>
            <CFormSelect id="inputState">
              <option>Choose...</option>
              <option>...</option>
            </CFormSelect>
          </CCol>
          <CCol :md="2">
            <CFormLabel for="inputZip">Zip</CFormLabel>
            <CFormInput id="inputZip" />
          </CCol>
          <CCol :xs="12">
            <CFormLabel for="gridCheck">Check me out</CFormLabel>
            <CFormCheck id="gridCheck" type="checkbox" />
          </CCol>
        </CForm>
      </CCol>
    </CRow>
  </CContainer>
</template>

<script>
export default {
  name: 'Plugins',
}
</script>
