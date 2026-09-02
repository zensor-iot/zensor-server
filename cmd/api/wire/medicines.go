//go:build wireinject
// +build wireinject

package wire

import (
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/config"

	controlPlaneUsecases "zensor-server/internal/control_plane/usecases"

	medicinesHTTPAPI "zensor-server/internal/medicines/httpapi"
	medicinesPersistence "zensor-server/internal/medicines/persistence"
	medicinesUsecases "zensor-server/internal/medicines/usecases"
	sharedPersistence "zensor-server/internal/shared_kernel/persistence"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	"github.com/google/wire"
)

func InitializeMedicinePatientController() (*medicinesHTTPAPI.PatientController, error) {
	wire.Build(
		provideAppConfig,
		medicinesPersistence.NewPatientRepository,
		wire.Bind(new(medicinesUsecases.PatientRepository), new(*medicinesPersistence.SimplePatientRepository)),
		medicinesPersistence.NewTreatmentRepository,
		wire.Bind(new(medicinesUsecases.TreatmentRepository), new(*medicinesPersistence.SimpleTreatmentRepository)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		DeviceServiceSet,
		wire.Bind(new(sharedUsecases.DeviceAdopter), new(*controlPlaneUsecases.SimpleDeviceService)),
		sharedUsecases.NewTenantService,
		wire.Bind(new(sharedUsecases.TenantService), new(*sharedUsecases.SimpleTenantService)),
		medicinesUsecases.NewPatientService,
		wire.Bind(new(medicinesUsecases.PatientService), new(*medicinesUsecases.SimplePatientService)),
		medicinesHTTPAPI.NewPatientController,
	)
	return nil, nil
}

func InitializeMedicineTreatmentController() (*medicinesHTTPAPI.TreatmentController, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		medicinesPersistence.NewTreatmentRepository,
		wire.Bind(new(medicinesUsecases.TreatmentRepository), new(*medicinesPersistence.SimpleTreatmentRepository)),
		medicinesPersistence.NewPatientRepository,
		wire.Bind(new(medicinesUsecases.PatientRepository), new(*medicinesPersistence.SimplePatientRepository)),
		medicinesPersistence.NewDoseRepository,
		wire.Bind(new(medicinesUsecases.DoseRepository), new(*medicinesPersistence.SimpleDoseRepository)),
		medicinesUsecases.NewTreatmentService,
		wire.Bind(new(medicinesUsecases.TreatmentService), new(*medicinesUsecases.SimpleTreatmentService)),
		medicinesHTTPAPI.NewTreatmentController,
	)
	return nil, nil
}

func InitializeMedicineDoseController() (*medicinesHTTPAPI.DoseController, error) {
	wire.Build(
		provideAppConfig,
		provideDatabase,
		medicinesPersistence.NewDoseRepository,
		wire.Bind(new(medicinesUsecases.DoseRepository), new(*medicinesPersistence.SimpleDoseRepository)),
		medicinesUsecases.NewDoseService,
		wire.Bind(new(medicinesUsecases.DoseService), new(*medicinesUsecases.SimpleDoseService)),
		medicinesHTTPAPI.NewDoseController,
	)
	return nil, nil
}

func InitializeMedicineWorker(broker async.InternalBroker) (*medicinesUsecases.MedicineWorker, error) {
	wire.Build(
		provideAppConfig,
		provideMedicineWorkerTicker,
		medicinesPersistence.NewTreatmentRepository,
		wire.Bind(new(medicinesUsecases.TreatmentRepository), new(*medicinesPersistence.SimpleTreatmentRepository)),
		medicinesPersistence.NewDoseRepository,
		wire.Bind(new(medicinesUsecases.DoseRepository), new(*medicinesPersistence.SimpleDoseRepository)),
		sharedPersistence.NewTenantRepository,
		wire.Bind(new(sharedUsecases.TenantRepository), new(*sharedPersistence.SimpleTenantRepository)),
		sharedPersistence.NewTenantConfigurationRepository,
		wire.Bind(new(sharedUsecases.TenantConfigurationRepository), new(*sharedPersistence.SimpleTenantConfigurationRepository)),
		sharedPersistence.NewUserRepository,
		wire.Bind(new(sharedUsecases.UserRepository), new(*sharedPersistence.SimpleUserRepository)),
		sharedUsecases.NewUserService,
		wire.Bind(new(sharedUsecases.UserService), new(*sharedUsecases.SimpleUserService)),
		DeviceServiceSet,
		wire.Bind(new(sharedUsecases.DeviceAdopter), new(*controlPlaneUsecases.SimpleDeviceService)),
		sharedUsecases.NewTenantService,
		wire.Bind(new(sharedUsecases.TenantService), new(*sharedUsecases.SimpleTenantService)),
		sharedUsecases.NewTenantConfigurationService,
		wire.Bind(new(sharedUsecases.TenantConfigurationService), new(*sharedUsecases.SimpleTenantConfigurationService)),
		medicinesUsecases.NewMedicineWorker,
	)
	return nil, nil
}

// provideMedicineWorkerTicker defaults to a minute. The reminder lead time in
// the worker has to stay strictly greater than this interval, or a dose can
// fall between two windows and its reminder is lost.
func provideMedicineWorkerTicker(appConfig config.AppConfig) *time.Ticker {
	interval := appConfig.Medicines.Worker.TickerInterval
	if interval == 0 {
		interval = time.Minute
	}
	return time.NewTicker(interval)
}
