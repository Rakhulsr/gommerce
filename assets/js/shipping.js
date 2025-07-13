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

document.addEventListener('DOMContentLoaded', function() {

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
        checkoutButton: document.querySelector('#calculate-shipping + button'),
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
    }

    /**
     
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
            updateGrandTotalDisplay(0);
        }
        setShippingMessage('');
    }

    /**
     *
     * @param {string} message
     * @param {'success'|'warning'|'error'|''} type
     */
    function setShippingMessage(message, type = '') {
        elements.shippingMessage.textContent = message;
        elements.shippingMessage.className = 'text-sm mt-2';
        // Hapus semua kelas warna sebelumnya
        elements.shippingMessage.classList.remove('text-red-500', 'text-yellow-500', 'text-green-500');
        if (type === 'error') {
            elements.shippingMessage.classList.add('text-red-500');
        } else if (type === 'warning') {
            elements.shippingMessage.classList.add('text-yellow-500');
        } else if (type === 'success') {
            elements.shippingMessage.classList.add('text-green-500');
        }
    }

    /**

     *
     * @param {Response} response
     * @param {string} errorMessagePrefix
     * @returns {Promise<Object>}
     * @throws {Error}
     */
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



    /**
    
     *
     * @param {string} provinceId
     */
    async function loadCities(provinceId) {
        resetShippingElements({ resetCity: true, resetCourier: false, resetShippingFee: true });
        elements.citySelect.innerHTML = '<option value="" selected>--Memuat Kota/Kabupaten--</option>';

        if (!provinceId) {
            setShippingMessage('Mohon pilih Provinsi terlebih dahulu.', 'warning');
            resetShippingElements({ resetCity: true, resetCourier: true, resetShippingFee: true }); // Reset kurir juga jika provinsi kosong
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
                elements.citySelect.disabled = false;

                elements.courierSelect.disabled = false;
                setShippingMessage('Silakan pilih Kota/Kabupaten dan Kurir.', 'warning');

                if (elements.citySelect.value && elements.courierSelect.value) {
                    loadShippingCosts();
                }

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

  
    async function loadShippingCosts() {
      
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
            const data = await handleApiResponse(await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    origin: originCityID,
                    destination: destinationCityID,
                    weight: totalWeight,
                    courier: courier
                })
            }), 'Gagal menghitung ongkir');

            if (data.costs && Array.isArray(data.costs)) {
                elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Pilih Opsi Pengiriman--</option>';
                const results = data.costs;

                if (results.length > 0 && results[0].costs && results[0].costs.length > 0) {
                    results[0].costs.forEach(cost => {
                        cost.cost.forEach(service => {
                            const option = document.createElement('option');
                            const formattedValue = `${results[0].code.toUpperCase()} - ${cost.service} (${service.etd} hari): ${formatCurrency(service.value)}`;
                            option.value = service.value;
                            option.textContent = formattedValue;
                            elements.shippingFeeSelect.appendChild(option);
                        });
                    });
                    elements.shippingFeeSelect.disabled = false;
                    setShippingMessage('');
                } else {
                    setShippingMessage('Tidak ada opsi pengiriman tersedia untuk rute ini atau kurir yang dipilih.', 'warning');
                    elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Tidak Tersedia--</option>';
                }
            } else {
                console.error('Backend response for shipping costs missing or invalid "costs" property:', data);
                throw new Error('Gagal menghitung ongkir: Data tidak lengkap atau format tidak sesuai.');
            }
        } catch (error) {
            console.error('Error in loadShippingCosts:', error);
            setShippingMessage(error.message || 'Terjadi kesalahan jaringan saat menghitung ongkir.', 'error');
            elements.shippingFeeSelect.innerHTML = '<option value="" selected>--Gagal Memuat--</option>';
        } finally {
            elements.shippingFeeSelect.disabled = false;
        }
    }

   
    resetShippingElements({ resetCity: true, resetCourier: true, resetShippingFee: true });
    updateGrandTotalDisplay(0);

    // Logika untuk mengisi dropdown berdasarkan nilai yang mungkin sudah ada (misal dari refresh halaman)
    if (elements.provinceSelect.value) {
        // Jika provinsi sudah dipilih, muat kota dan kemudian cek kondisi kurir/ongkir
        loadCities(elements.provinceSelect.value).then(() => {
            if (elements.citySelect.value && elements.courierSelect.value) {
                loadShippingCosts();
            } else if (elements.citySelect.value) {
                setShippingMessage('Silakan pilih Kurir.', 'warning');
            } else {
                setShippingMessage('Silakan pilih Kota/Kabupaten.', 'warning');
            }
        });
    } else {
        // Jika tidak ada provinsi yang terpilih, pastikan kurir dan ongkir disabled.
        elements.courierSelect.disabled = true;
        elements.shippingFeeSelect.disabled = true;
        setShippingMessage('Mohon pilih Provinsi terlebih dahulu untuk melihat opsi pengiriman.', 'warning');
    }

    // --- 5. EVENT LISTENERS ---

    elements.provinceSelect.addEventListener('change', function() {
        const selectedProvinceId = this.value;

        // Reset semua elemen terkait pengiriman: kota, kurir, biaya pengiriman.
        // Opsi kurir akan tetap ada di DOM, tetapi nilai terpilihnya di-reset dan elemennya di-disable.
        resetShippingElements({ resetCity: true, resetCourier: true, resetShippingFee: true });
        loadCities(selectedProvinceId); // Ini akan mengaktifkan kembali courierSelect jika berhasil

        if (!selectedProvinceId) {
            setShippingMessage('Mohon pilih Provinsi terlebih dahulu untuk melihat opsi pengiriman.', 'warning');
            elements.courierSelect.disabled = true;
            elements.shippingFeeSelect.disabled = true;
        }
    });

    elements.citySelect.addEventListener('change', function() {
        const selectedCityId = this.value;
        if (selectedCityId) {
            elements.courierSelect.disabled = false; // Aktifkan kurir jika kota dipilih
            if (elements.courierSelect.value) {
                loadShippingCosts();
            } else {
                setShippingMessage('Mohon pilih Kurir.', 'warning');
                resetShippingElements({ resetCity: false, resetCourier: false, resetShippingFee: true });
            }
        } else {
            // Kota tidak terpilih, reset kurir dan biaya pengiriman.
            resetShippingElements({ resetCity: false, resetCourier: true, resetShippingFee: true });
            setShippingMessage('Mohon pilih Kota/Kabupaten.', 'warning');
        }
    });

    elements.courierSelect.addEventListener('change', function() {
        const selectedCourier = this.value;
        if (selectedCourier && elements.citySelect.value) {
            loadShippingCosts();
        } else {
            resetShippingElements({ resetCity: false, resetCourier: false, resetShippingFee: true });
            if (!selectedCourier) {
                setShippingMessage('Mohon pilih Kurir.', 'warning');
            } else if (!elements.citySelect.value) {
                setShippingMessage('Mohon pilih Kota/Kabupaten terlebih dahulu.', 'warning');
            }
        }
    });

    elements.shippingFeeSelect.addEventListener('change', function() {
        const selectedValue = this.value;
        updateGrandTotalDisplay(selectedValue);

        if (selectedValue === "") {
            setShippingMessage('Silakan pilih opsi pengiriman.', 'warning');
        } else {
            setShippingMessage('');
        }
    });

    elements.checkoutButton.addEventListener('click', function(event) {
        event.preventDefault();

        const province = elements.provinceSelect.value;
        const city = elements.citySelect.value;
        const courier = elements.courierSelect.value;
        const shippingFee = elements.shippingFeeSelect.value;

        if (!province) {
            setShippingMessage('Mohon pilih Provinsi sebelum melanjutkan.', 'error');
            return;
        }
        if (!city) {
            setShippingMessage('Mohon pilih Kota/Kabupaten sebelum melanjutkan.', 'error');
            return;
        }
        if (!courier) {
            setShippingMessage('Mohon pilih Kurir sebelum melanjutkan.', 'error');
            return;
        }
        if (!shippingFee) {
            setShippingMessage('Mohon pilih Opsi Pengiriman sebelum melanjutkan.', 'error');
            return;
        }

        console.log('Semua input pengiriman valid. Siap untuk Checkout!');
        setShippingMessage('Semua siap untuk checkout!', 'success');
        // If all validation passes, you can proceed with form submission or navigation.
        // elements.calculateShippingForm.submit(); // Anda bisa mengaktifkan ini jika ini adalah form submit
        // window.location.href = '/checkout-page'; // Atau ini untuk navigasi
    });

    document.querySelectorAll('.delete-confirm-form').forEach(form => {
        form.addEventListener('submit', function(event) {
            if (!confirm('Apakah Anda yakin ingin menghapus item ini dari keranjang?')) {
                event.preventDefault();
            }
        });
    });

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