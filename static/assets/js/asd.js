/**
 * @param {number|string} amount
 * @returns {string}
 */
function formatCurrency(amount) {
    const num = parseInt(amount, 10);
    if (isNaN(num)) {
        return "Rp 0";
    }
    return `Rp ${num.toLocaleString('id-ID')}`;
}

// Generic API response handler
async function handleApiResponse(response, errorMessagePrefix) {
    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
        const errorText = await response.text();
        console.error(`Expected JSON response, but received:`, errorText);
        throw new Error(`${errorMessagePrefix}: Respons tidak valid atau bukan JSON.`);
    }

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        console.error(`HTTP Error:`, response.status, errorData);
        let errorMessage = `${errorMessagePrefix}: ${errorData.message || response.statusText}`;
        if (errorData && errorData.error) {
            errorMessage = `${errorMessagePrefix}: ${errorData.error}`;
        } else if (errorData && errorData.rajaongkir && errorData.rajaongkir.status && errorData.rajaongkir.status.description) {
            errorMessage = `${errorMessagePrefix}: ${errorData.rajaongkir.status.description}`;
        }
        throw new Error(errorMessage);
    }
    return response.json();
}


document.addEventListener('DOMContentLoaded', function() { 
    // SweetAlert2 for delete confirmation
    document.querySelectorAll('.delete-cart-item-form').forEach(form => {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            const formElement = this; 

            Swal.fire({
                title: 'Apakah Anda yakin?',
                text: 'Item ini akan dihapus dari keranjang!',
                icon: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#3085d6',
                cancelButtonColor: '#d33',
                confirmButtonText: 'Ya, hapus!',
                cancelButtonText: 'Batal'
            }).then((result) => {
                if (result.isConfirmed) {
                    formElement.submit();
                }
            });
        });
    });

    const elements = {
        provinceSelect: document.getElementById('province_id'),
        citySelect: document.getElementById('city_id'),
        courierSelect: document.getElementById('courier'),
        shippingFeeSelect: document.getElementById('shipping_fee'),
        calculateShippingForm: document.getElementById('calculate-shipping'),
        shippingMessage: document.getElementById('shiping-calculation-msg'),
        shippingFeeDisplay: document.getElementById('shipping-fee-display'),
        grandTotalDisplay: document.getElementById('grand-total-display'),
        cartSubtotalInput: document.getElementById('cart_subtotal_input'),
        totalWeightInput: document.getElementById('total_weight_input'),
        originCityIDInput: document.getElementById('origin_city_id_input'),
        
        addressSelect: document.getElementById('address_select'), 
        proceedToCheckoutBtn: document.getElementById('proceedToCheckoutBtn'), 
        checkoutShippingCostInput: document.getElementById('checkout_shipping_cost'), 
        checkoutShippingServiceCodeInput: document.getElementById('checkout_shipping_service_code'), 
        checkoutShippingServiceNameInput: document.getElementById('checkout_shipping_service_name'), 
        checkoutSelectedAddressIdInput: document.getElementById('checkout_selected_address_id'), 
        checkoutFinalTotalPriceInput: document.getElementById('checkout_final_total_price'), // NEW: for final price
    };

    let totalWeight = 1;
    if (elements.totalWeightInput && !isNaN(parseInt(elements.totalWeightInput.value, 10))) {
        totalWeight = parseInt(elements.totalWeightInput.value, 10);
        if (totalWeight <= 0) {
            totalWeight = 1;
        }
    }

    const originCityID = elements.originCityIDInput ? String(elements.originCityIDInput.value) : "";
    const APP_API_BASE_URL = "/";

    let cartSubtotal = 0;
    if (elements.cartSubtotalInput && !isNaN(parseInt(elements.cartSubtotalInput.value, 10))) {
        cartSubtotal = parseInt(elements.cartSubtotalInput.value, 10);
    }

    /**
     * @param {string} selectedShippingCostValue
     */
    function updateGrandTotalDisplay(selectedShippingCostValue = "0") {
        const currentShippingFee = parseInt(selectedShippingCostValue, 10) || 0;
        const grandTotalCalculated = cartSubtotal + currentShippingFee;

        elements.shippingFeeDisplay.textContent = formatCurrency(currentShippingFee);
        elements.grandTotalDisplay.textContent = formatCurrency(grandTotalCalculated);
        elements.checkoutFinalTotalPriceInput.value = grandTotalCalculated; 

        
        const isAddressSelected = elements.addressSelect.value !== "";
        const isShippingSelected = elements.shippingFeeSelect.value !== "";
        console.log(`[updateGrandTotalDisplay] Current shippingFeeSelect.value: '${elements.shippingFeeSelect.value}'`);
        console.log(`[updateGrandTotalDisplay] isAddressSelected: ${isAddressSelected}, isShippingSelected: ${isShippingSelected}`);
        elements.proceedToCheckoutBtn.disabled = !(isAddressSelected && isShippingSelected);
        console.log(`[updateGrandTotalDisplay] Proceed to Checkout Button Disabled: ${elements.proceedToCheckoutBtn.disabled}`);
    }

    /**
     * Resets shipping related dropdowns to their default disabled state, except for province.
     * @param {Object} options
     * @param {boolean} [options.resetCity=true]
     * @param {boolean} [options.resetCourier=true]
     * @param {boolean} [options.resetShippingFee=true]
     * @param {string} [options.cityDefaultText='--Pilih Kota/Kabupaten--']
     * @param {string} [options.courierDefaultText='--Pilih Kurir--']
     * @param {string} [options.shippingFeeDefaultText='--Pilih Opsi Pengiriman--']
     */
    function resetShippingElements({
        resetCity = true,
        resetCourier = true,
        resetShippingFee = true,
        cityDefaultText = '--Pilih Kota/Kabupaten--',
        courierDefaultText = '--Pilih Kurir--',
        shippingFeeDefaultText = '--Pilih Opsi Pengiriman--'
    } = {}) {
        console.log(`[resetShippingElements] Resetting options: resetCity=${resetCity}, resetCourier=${resetCourier}, resetShippingFee=${resetShippingFee}`);

        if (resetCity) {
            elements.citySelect.innerHTML = `<option value="" selected>${cityDefaultText}</option>`;
            elements.citySelect.disabled = true;
        }
        if (resetCourier) {
            elements.courierSelect.value = "";
            elements.courierSelect.disabled = true;
        }
        if (resetShippingFee) {
            elements.shippingFeeSelect.innerHTML = `<option value="" selected>${shippingFeeDefaultText}</option>`;
            elements.shippingFeeSelect.disabled = true;
            updateGrandTotalDisplay(0); // Update grand total and re-evaluate checkout button status
            elements.checkoutShippingCostInput.value = 0; 
            elements.checkoutShippingServiceCodeInput.value = "";
            elements.checkoutShippingServiceNameInput.value = "";
            elements.checkoutFinalTotalPriceInput.value = cartSubtotal; // Reset final price to just cart subtotal
        }
        setShippingMessage('');
    }

    /**
     * @param {string} message
     * @param {'success'|'warning'|'error'|''} type
     */
    function setShippingMessage(message, type = '') {
        elements.shippingMessage.textContent = message;
        elements.shippingMessage.className = 'text-sm mt-2';
        
        elements.shippingMessage.classList.remove('text-red-500', 'text-yellow-500', 'text-green-500');
        if (type === 'error') {
            elements.shippingMessage.classList.add('text-red-500');
        } else if (type === 'warning') {
            elements.shippingMessage.classList.add('text-yellow-500');
        } else if (type === 'success') {
            elements.shippingMessage.classList.add('text-green-500');
        }
        console.log(`[setShippingMessage] Type: ${type}, Message: ${message}`);
    }

    /**
     * @param {string} provinceId
     * @param {string} [selectedCityIdToPreselect=""]
     */
    async function loadCities(provinceId, selectedCityIdToPreselect = "") {
        console.log(`[loadCities] Loading cities for province: ${provinceId}, pre-select city: ${selectedCityIdToPreselect}`);
        // Reset city, courier, shipping fee dropdowns
        resetShippingElements({ resetCity: true, resetCourier: true, resetShippingFee: true }); 
        elements.citySelect.innerHTML = '<option value="" selected>--Memuat Kota/Kabupaten--</option>';

        if (!provinceId) {
            setShippingMessage('Mohon pilih Provinsi terlebih dahulu.', 'warning');
            return;
        }

        try {
            const url = `${APP_API_BASE_URL}cities?province_id=${provinceId}`;
            const data = await handleApiResponse(await fetch(url, { method: 'GET' }), 'Gagal memuat kota');

            if (data.cities && Array.isArray(data.cities)) {
                elements.citySelect.innerHTML = '<option value="" selected>--Pilih Kota/Kabupaten--</option>';
                data.cities.forEach(city => {
                    const option = document.createElement('option');
                    option.value = city.city_id;
                    option.textContent = `${city.type} ${city.city_name}`;
                    elements.citySelect.appendChild(option);
                });
                elements.citySelect.disabled = false; // Enable City dropdown
                console.log('[loadCities] City dropdown enabled.');
                
                // Pre-select the city if provided
                if (selectedCityIdToPreselect) {
                    elements.citySelect.value = selectedCityIdToPreselect;
                    console.log(`[loadCities] Pre-selected city: ${selectedCityIdToPreselect}`);
                }

                // Courier select should be enabled after city is selected
                elements.courierSelect.disabled = false; 
                console.log('[loadCities] Courier dropdown enabled.');

                setShippingMessage('Silakan pilih Kota/Kabupaten dan Kurir.', 'warning');

            } else {
                console.error('Backend response for cities missing or invalid "cities" property:', data);
                throw new Error('Gagal memuat kota: Data tidak lengkap atau format tidak sesuai.');
            }
        } catch (error) {
            console.error('Error in loadCities:', error);
            setShippingMessage(error.message || 'Terjadi kesalahan jaringan saat memuat kota.', 'error');
            elements.citySelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
            elements.citySelect.disabled = true;
            elements.courierSelect.disabled = true; 
            elements.shippingFeeSelect.disabled = true;
        }
    }

    /**
     * @param {string} [selectedProvinceIdToPreselect=""]
     */
    async function loadProvincesForDestination(selectedProvinceIdToPreselect = "") {
        console.log(`[loadProvincesForDestination] Loading provinces, pre-select province: ${selectedProvinceIdToPreselect}`);
        elements.provinceSelect.innerHTML = '<option value="" selected>--Memuat Provinsi--</option>';
        elements.provinceSelect.disabled = true; // Temporarily disable while loading

        try {
            const url = `${APP_API_BASE_URL}provinces`;
            const data = await handleApiResponse(await fetch(url, { method: 'GET' }), 'Gagal memuat provinsi');

            if (data.provinces && Array.isArray(data.provinces)) {
                elements.provinceSelect.innerHTML = '<option value="" selected>--Pilih Provinsi--</option>';
                data.provinces.forEach(province => {
                    const option = document.createElement('option');
                    option.value = province.province_id;
                    option.textContent = province.province_name;
                    elements.provinceSelect.appendChild(option);
                });
                elements.provinceSelect.disabled = false; // Enable Province dropdown
                console.log('[loadProvincesForDestination] Province dropdown enabled.');
                
                if (selectedProvinceIdToPreselect) {
                    elements.provinceSelect.value = selectedProvinceIdToPreselect;
                    console.log(`[loadProvincesForDestination] Pre-selected province: ${selectedProvinceIdToPreselect}`);
                }
                setShippingMessage('Pilih Provinsi, Kota, dan Kurir untuk menghitung ongkir.', 'warning');
            } else {
                console.error('Backend response for provinces missing or invalid "provinces" property:', data);
                throw new Error('Gagal memuat provinsi: Data tidak lengkap atau format tidak sesuai.');
            }
        } catch (error) {
            console.error('Error in loadProvincesForDestination:', error);
            setShippingMessage(error.message || 'Terjadi kesalahan jaringan saat memuat provinsi.', 'error');
            elements.provinceSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
            elements.provinceSelect.disabled = true;
        }
    }
 
    async function loadShippingCosts() {
        console.log('[loadShippingCosts] Loading shipping costs...');
        // Only reset shipping fee. Keep city and courier as is.
        resetShippingElements({ resetCity: false, resetCourier: false, resetShippingFee: true }); 
        elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Memuat Opsi Pengiriman--</option>';

        const destinationCityID = elements.citySelect.value;
        const courier = elements.courierSelect.value;

        if (!originCityID) {
            setShippingMessage('Kota asal pengiriman belum ditentukan.', 'error');
            elements.shippingFeeSelect.disabled = true;
            return;
        }
        if (totalWeight <= 0) {
            setShippingMessage('Berat total produk harus lebih dari 0.', 'error');
            elements.shippingFeeSelect.disabled = true;
            return;
        }
        if (!destinationCityID) { 
            setShippingMessage('Mohon pilih Kota/Kabupaten.', 'warning');
            elements.shippingFeeSelect.disabled = true;
            return;
        }
        if (!courier) { 
            setShippingMessage('Mohon pilih Kurir.', 'warning');
            elements.shippingFeeSelect.disabled = true;
            return;
        }

        setShippingMessage('Memuat opsi pengiriman...', 'warning');

        try {
            const url = `${APP_API_BASE_URL}calculate-shipping-cost`;
            // TIDAK MENGGUNAKAN CSRF TOKEN SESUAI PERMINTAAN ANDA
            const data = await handleApiResponse(await fetch(url, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    origin: originCityID,
                    destination: destinationCityID,
                    weight: totalWeight,
                    courier: courier
                })
            }), 'Gagal menghitung ongkir');

            if (data.results && Array.isArray(data.results)) { // Changed from data.costs to data.results as per common API response
                elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Pilih Opsi Pengiriman--</option>';
                const results = data.results;

                if (results.length > 0 && results[0].costs && results[0].costs.length > 0) {
                    results[0].costs.forEach(cost => {
                        cost.cost.forEach(service => {
                            const option = document.createElement('option');
                            const formattedValue = `${results[0].code.toUpperCase()} - ${cost.service} (${service.etd} hari): ${formatCurrency(service.value)}`;
                            option.value = service.value;
                            option.textContent = formattedValue;
                            option.dataset.serviceCode = results[0].code; // Store service code
                            option.dataset.serviceName = `${results[0].name} - ${cost.service}`; // Store service name
                            elements.shippingFeeSelect.appendChild(option);
                        });
                    });
                    elements.shippingFeeSelect.disabled = false; // Enable Shipping Fee dropdown
                    console.log('[loadShippingCosts] Shipping options loaded and dropdown enabled.');
                    setShippingMessage('');
                } else {
                    setShippingMessage('Tidak ada opsi pengiriman tersedia untuk rute ini atau kurir yang dipilih.', 'warning');
                    elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Tidak Tersedia--</option>';
                    elements.shippingFeeSelect.disabled = true; // Ensure disabled if no options
                    console.log('[loadShippingCosts] No shipping options available.');
                }
            } else {
                console.error('Backend response for shipping costs missing or invalid "results" property:', data);
                throw new Error('Gagal menghitung ongkir: Data tidak lengkap atau format tidak sesuai.');
            }
        } catch (error) {
            console.error('Error in loadShippingCosts:', error);
            setShippingMessage(error.message || 'Terjadi kesalahan jaringan saat menghitung ongkir.', 'error');
            elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
            elements.shippingFeeSelect.disabled = true; // Ensure disabled on error
        } finally {
            // No need to enable here, as it's handled in success/failure branches
        }
    }

    // --- EVENT LISTENERS ---

    // Initial setup when page loads
    // Explicitly disable city, courier, shipping fee dropdowns at start
    elements.citySelect.disabled = true;
    elements.courierSelect.disabled = true;
    elements.shippingFeeSelect.disabled = true;
    updateGrandTotalDisplay(0); // Set initial grand total and disable checkout button

    loadProvincesForDestination(); // Load provinces initially. This will enable provinceSelect.

    // Initial check if an address is already selected on page load (e.g., if it's the primary address)
    // This part should NOT trigger full shipping calculation, only update address ID and button status.
    const initialAddressSelectValue = elements.addressSelect.value;
    if (initialAddressSelectValue !== "") {
        console.log(`[DOMContentLoaded] Initial address selected: ${initialAddressSelectValue}. Updating checkout address ID.`);
        elements.checkoutSelectedAddressIdInput.value = initialAddressSelectValue;
        // Just update grand total display to re-evaluate button state based on address selection
        updateGrandTotalDisplay(elements.shippingFeeSelect.value); 
    } else {
        console.log('[DOMContentLoaded] No initial address selected. Checkout button remains disabled.');
    }

    // Event listener for address selection
    elements.addressSelect.addEventListener('change', async function() {
        console.log('[addressSelect] Change event triggered.');
        const selectedAddressId = this.value;

        elements.checkoutSelectedAddressIdInput.value = selectedAddressId; 
        console.log(`[addressSelect] Selected Address ID: ${selectedAddressId}`);

        // ONLY update the checkout button status. Do NOT reset shipping dropdowns or trigger their loads.
        updateGrandTotalDisplay(elements.shippingFeeSelect.value); 

        if (!selectedAddressId) {
            setShippingMessage('Pilih alamat pengiriman untuk melihat opsi ongkir.', 'warning');
            // If address is unselected, clear the hidden address ID
            elements.checkoutSelectedAddressIdInput.value = "";
        } else {
            // If an address is selected, clear the message if it was about selecting an address
            setShippingMessage('');
        }
    });

    elements.provinceSelect.addEventListener('change', function() {
        console.log('[provinceSelect] Change event triggered.');
        const selectedProvinceId = this.value;
        // Reset city, courier, shipping fee dropdowns
        resetShippingElements({ resetCity: true, resetCourier: true, resetShippingFee: true });
        loadCities(selectedProvinceId); // Load cities based on selected province

        if (!selectedProvinceId) {
            setShippingMessage('Mohon pilih Provinsi terlebih dahulu untuk melihat opsi pengiriman.', 'warning');
            elements.courierSelect.disabled = true;
            elements.shippingFeeSelect.disabled = true;
        }
    });

    elements.citySelect.addEventListener('change', function() {
        console.log('[citySelect] Change event triggered.');
        const selectedCityId = this.value;
        console.log(`[citySelect] Selected City ID: ${selectedCityId}`);
        if (selectedCityId) {
            elements.courierSelect.disabled = false; // Enable Courier dropdown when city is selected
            console.log('[citySelect] City selected, enabling courier dropdown.');
            if (elements.courierSelect.value) {
                console.log('[citySelect] City and courier selected, loading shipping costs.');
                loadShippingCosts();
            } else {
                setShippingMessage('Mohon pilih Kurir.', 'warning');
                // Only reset shipping fee, keep city and courier as is.
                resetShippingElements({ resetCity: false, resetCourier: false, resetShippingFee: true });
            }
        } else {
            // If city is unselected, reset courier and shipping fee.
            resetShippingElements({ resetCity: false, resetCourier: true, resetShippingFee: true });
            setShippingMessage('Mohon pilih Kota/Kabupaten.', 'warning');
        }
    });

    elements.courierSelect.addEventListener('change', function() {
        console.log('[courierSelect] Change event triggered.');
        const selectedCourier = this.value;
        console.log(`[courierSelect] Selected Courier: ${selectedCourier}`);
        if (selectedCourier && elements.citySelect.value) {
            console.log('[courierSelect] Courier and city selected, loading shipping costs.');
            loadShippingCosts();
        } else {
            // Only reset shipping fee. Keep city and courier as is.
            resetShippingElements({ resetCity: false, resetCourier: false, resetShippingFee: true });
            if (!selectedCourier) {
                setShippingMessage('Mohon pilih Kurir.', 'warning');
            } else if (!elements.citySelect.value) {
                setShippingMessage('Mohon pilih Kota/Kabupaten terlebih dahulu.', 'warning');
            }
        }
    });

    elements.shippingFeeSelect.addEventListener('change', function() {
        console.log('[shippingFeeSelect] Change event triggered.');
        const selectedOption = this.options[this.selectedIndex];
        const selectedValue = this.value;
        console.log(`[shippingFeeSelect] Selected Shipping Cost Value: '${selectedValue}'`); // Log the actual value
        updateGrandTotalDisplay(selectedValue); // This will also update the checkout button status

        // Populate hidden inputs for the checkout form
        elements.checkoutShippingCostInput.value = selectedValue;
        elements.checkoutShippingServiceCodeInput.value = selectedOption.dataset.serviceCode || "";
        elements.checkoutShippingServiceNameInput.value = selectedOption.dataset.serviceName || "";
        console.log(`[shippingFeeSelect] Hidden inputs updated: Cost=${elements.checkoutShippingCostInput.value}, Code=${elements.checkoutShippingServiceCodeInput.value}, Name=${elements.checkoutShippingServiceNameInput.value}`);

        if (selectedValue === "") {
            setShippingMessage('Silakan pilih opsi pengiriman.', 'warning');
        } else {
            setShippingMessage('');
        }
    });

    // Event listener for the actual checkout button (outside calculate-shipping form)
    elements.proceedToCheckoutBtn.addEventListener('click', function(event) {
        console.log('[proceedToCheckoutBtn] Click event triggered.');
        // Prevent default form submission initially to perform client-side validation
        event.preventDefault(); 

        const addressSelected = elements.addressSelect.value;
        const provinceSelected = elements.provinceSelect.value; // These are from the shipping calculation form
        const citySelected = elements.citySelect.value;       // These are from the shipping calculation form
        const courierSelected = elements.courierSelect.value; // These are from the shipping calculation form
        const shippingFeeSelected = elements.shippingFeeSelect.value; // These are from the shipping calculation form

        console.log(`[proceedToCheckoutBtn] Validation check:
            Address: '${addressSelected}'
            Province (Shipping): '${provinceSelected}'
            City (Shipping): '${citySelected}'
            Courier (Shipping): '${courierSelected}'
            Shipping Fee (Selected): '${shippingFeeSelected}'`);

        if (!addressSelected) {
            setShippingMessage('Mohon pilih Alamat Pengiriman sebelum melanjutkan.', 'error');
            return;
        }
        // Validate shipping calculation fields only if an address is selected and they are relevant
        // The button disabled state already handles if shipping is not selected.
        if (!provinceSelected || !citySelected || !courierSelected || !shippingFeeSelected) {
             setShippingMessage('Mohon lengkapi semua opsi pengiriman (Provinsi, Kota, Kurir, Opsi Pengiriman) sebelum melanjutkan.', 'error');
             return;
        }


        // If all validations pass, manually submit the checkout form
        console.log('Semua input pengiriman dan alamat valid. Siap untuk Checkout! Submitting form...');
        setShippingMessage('Semua siap untuk checkout!', 'success');
        this.closest('form').submit(); // Submit the parent form of the button
    });

    // Flash message dismissal
    const updateSuccessNotification = document.getElementById('update-success-notification');
    if (updateSuccessNotification) {
        setTimeout(() => {
            updateSuccessNotification.classList.add('animate-fade-out');
            updateSuccessNotification.addEventListener('animationend', () => {
                updateSuccessNotification.remove();
            });
        }, 5000);
    }
});