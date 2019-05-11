// +build windows

package fs

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

type HRESULT uint

const (
	S_OK                                      HRESULT = 0x00000000
	E_ACCESSDENIED                                    = 0x80070005
	E_OUTOFMEMORY                                     = 0x8007000E
	E_INVALIDARG                                      = 0x80070057
	VSS_E_BAD_STATE                                   = 0x80042301
	VSS_E_LEGACY_PROVIDER                             = 0x800423F7
	VSS_E_RESYNC_IN_PROGRESS                          = 0x800423FF
	VSS_E_SNAPSHOT_NOT_IN_SET                         = 0x8004232B
	VSS_E_MAXIMUM_NUMBER_OF_VOLUMES_REACHED           = 0x80042312
	VSS_E_MAXIMUM_NUMBER_OF_SNAPSHOTS_REACHED         = 0x80042317
	VSS_E_NESTED_VOLUME_LIMIT                         = 0x8004232C
	VSS_E_OBJECT_NOT_FOUND                            = 0x80042308
	VSS_E_PROVIDER_NOT_REGISTERED                     = 0x80042304
	VSS_E_PROVIDER_VETO                               = 0x80042306
	VSS_E_VOLUME_NOT_SUPPORTED                        = 0x8004230C
	VSS_E_VOLUME_NOT_SUPPORTED_BY_PROVIDER            = 0x8004230E
	VSS_E_UNEXPECTED                                  = 0x80042302
	VSS_E_UNEXPECTED_PROVIDER_ERROR                   = 0x8004230F
	VSS_E_UNSELECTED_VOLUME                           = 0x8004232A
	VSS_E_CANNOT_REVERT_DISKID                        = 0x800423FE
	VSS_E_INVALID_XML_DOCUMENT                        = 0x80042311
	VSS_E_OBJECT_ALREADY_EXISTS                       = 0x8004230D
	FSRVP_E_SHADOW_COPY_SET_IN_PROGRESS               = 0x80042316
)

var hresultToString = map[HRESULT]string{
	S_OK:                                    "S_OK",
	E_ACCESSDENIED:                          "E_ACCESSDENIED",
	E_OUTOFMEMORY:                           "E_OUTOFMEMORY",
	E_INVALIDARG:                            "E_INVALIDARG",
	VSS_E_BAD_STATE:                         "VSS_E_BAD_STATE",
	VSS_E_LEGACY_PROVIDER:                   "VSS_E_LEGACY_PROVIDER",
	VSS_E_RESYNC_IN_PROGRESS:                "VSS_E_RESYNC_IN_PROGRESS",
	VSS_E_SNAPSHOT_NOT_IN_SET:               "VSS_E_SNAPSHOT_NOT_IN_SET",
	VSS_E_MAXIMUM_NUMBER_OF_VOLUMES_REACHED: "VSS_E_MAXIMUM_NUMBER_OF_VOLUMES_REACHED",
	VSS_E_MAXIMUM_NUMBER_OF_SNAPSHOTS_REACHED: "VSS_E_MAXIMUM_NUMBER_OF_SNAPSHOTS_REACHED",
	VSS_E_NESTED_VOLUME_LIMIT:                 "VSS_E_NESTED_VOLUME_LIMIT",
	VSS_E_OBJECT_NOT_FOUND:                    "VSS_E_OBJECT_NOT_FOUND",
	VSS_E_PROVIDER_NOT_REGISTERED:             "VSS_E_PROVIDER_NOT_REGISTERED",
	VSS_E_PROVIDER_VETO:                       "VSS_E_PROVIDER_VETO",
	VSS_E_VOLUME_NOT_SUPPORTED:                "VSS_E_VOLUME_NOT_SUPPORTED",
	VSS_E_VOLUME_NOT_SUPPORTED_BY_PROVIDER:    "VSS_E_VOLUME_NOT_SUPPORTED_BY_PROVIDER",
	VSS_E_UNEXPECTED:                          "VSS_E_UNEXPECTED",
	VSS_E_UNEXPECTED_PROVIDER_ERROR:           "VSS_E_UNEXPECTED_PROVIDER_ERROR",
	VSS_E_UNSELECTED_VOLUME:                   "VSS_E_UNSELECTED_VOLUME",
	VSS_E_CANNOT_REVERT_DISKID:                "VSS_E_CANNOT_REVERT_DISKID",
	VSS_E_INVALID_XML_DOCUMENT:                "VSS_E_INVALID_XML_DOCUMENT",
	VSS_E_OBJECT_ALREADY_EXISTS:               "VSS_E_OBJECT_ALREADY_EXISTS",
	FSRVP_E_SHADOW_COPY_SET_IN_PROGRESS:       "FSRVP_E_SHADOW_COPY_SET_IN_PROGRESS",
}

func (h HRESULT) Str() string {
	if i, ok := hresultToString[h]; ok {
		return i
	} else {
		return "UNKNOWN"
	}
}

type VssContext int

const (
	VSS_CTX_BACKUP VssContext = iota
	VSS_CTX_FILE_SHARE_BACKUP
	VSS_CTX_NAS_ROLLBACK
	VSS_CTX_APP_ROLLBACK
	VSS_CTX_CLIENT_ACCESSIBLE
	VSS_CTX_CLIENT_ACCESSIBLE_WRITERS
	VSS_CTX_ALL
)

type VssBackup int

const (
	VSS_BT_UNDEFINED VssBackup = iota
	VSS_BT_FULL
	VSS_BT_INCREMENTAL
	VSS_BT_DIFFERENTIAL
	VSS_BT_LOG
	VSS_BT_COPY
	VSS_BT_OTHER
)

type VssObjectType int

const (
	VSS_OBJECT_UNKNOWN VssObjectType = iota
	VSS_OBJECT_NONE
	VSS_OBJECT_SNAPSHOT_SET
	VSS_OBJECT_SNAPSHOT
	VSS_OBJECT_PROVIDER
	VSS_OBJECT_TYPE_COUNT
)

var UUID_IVSS = &ole.GUID{
	0x665c1d5f, 0xc218, 0x414d,
	[8]byte{0xa0, 0x5d, 0x7f, 0xef, 0x5f, 0x9d, 0x5c, 0x86},
}

type IVssBackupComponents struct {
	ole.IUnknown
}
type IVssBackupComponentsVTable struct {
	ole.IUnknownVtbl
	getWriterComponentsCount      uintptr
	getWriterComponents           uintptr
	initializeForBackup           uintptr
	setBackupState                uintptr
	initializeForRestore          uintptr
	setRestoreState               uintptr
	gatherWriterMetadata          uintptr
	getWriterMetadataCount        uintptr
	getWriterMetadata             uintptr
	freeWriterMetadata            uintptr
	addComponent                  uintptr
	prepareForBackup              uintptr
	abortBackup                   uintptr
	gatherWriterStatus            uintptr
	getWriterStatusCount          uintptr
	freeWriterStatus              uintptr
	getWriterStatus               uintptr
	setBackupSucceeded            uintptr
	setBackupOptions              uintptr
	setSelectedForRestore         uintptr
	setRestoreOptions             uintptr
	setAdditionalRestores         uintptr
	setPreviousBackupStamp        uintptr
	saveAsXML                     uintptr
	backupComplete                uintptr
	addAlternativeLocationMapping uintptr
	addRestoreSubcomponent        uintptr
	setFileRestoreStatus          uintptr
	addNewTarget                  uintptr
	setRangesFilePath             uintptr
	preRestore                    uintptr
	postRestore                   uintptr
	setContext                    uintptr
	startSnapshotSet              uintptr
	addToSnapshotSet              uintptr
	doSnapshotSet                 uintptr
	deleteSnapshots               uintptr
	importSnapshots               uintptr
	breakSnapshotSet              uintptr
	getSnapshotProperties         uintptr
	query                         uintptr
	isVolumeSupported             uintptr
	disableWriterClasses          uintptr
	enableWriterClasses           uintptr
	disableWriterInstances        uintptr
	exposeSnapshot                uintptr
	revertToSnapshot              uintptr
	queryRevertStatus             uintptr
}

func (vss *IVssBackupComponents) getVTable() *IVssBackupComponentsVTable {
	return (*IVssBackupComponentsVTable)(unsafe.Pointer(vss.RawVTable))
}

func (vss *IVssBackupComponents) InitializeForBackup() HRESULT {
	result, _, _ := syscall.Syscall(
		vss.getVTable().initializeForBackup, 2,
		uintptr(unsafe.Pointer(vss)),
		0, 0,
	)

	return HRESULT(result)
}

func (vss *IVssBackupComponents) SetContext(context VssContext) HRESULT {
	result, _, _ := syscall.Syscall(
		vss.getVTable().setContext, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(context),
		0,
	)

	return HRESULT(result)
}

func (vss *IVssBackupComponents) GatherWriterMetadata() (HRESULT, *IVSSAsync) {
	var oleIUnknown *ole.IUnknown

	result, _, _ := syscall.Syscall(
		vss.getVTable().gatherWriterMetadata, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&oleIUnknown)),
		0,
	)

	return vss.handleAsyncReturnValue(result, oleIUnknown)
}

func (vss *IVssBackupComponents) StartSnapshotSet() (HRESULT, ole.GUID) {
	var snapshotSetId ole.GUID
	result, _, _ := syscall.Syscall(
		vss.getVTable().startSnapshotSet, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&snapshotSetId)),
		0,
	)
	return HRESULT(result), snapshotSetId
}

func (vss *IVssBackupComponents) AddToSnapshotSet(volumeName string, idSnapshot *ole.GUID) HRESULT {

	volumeNamePointer, _ := syscall.UTF16PtrFromString(volumeName)

	result, _, _ := syscall.Syscall6(
		vss.getVTable().addToSnapshotSet, 4,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(volumeNamePointer)),
		uintptr(unsafe.Pointer(ole.IID_NULL)),
		uintptr(unsafe.Pointer(idSnapshot)),
		0, 0,
	)

	return HRESULT(result)
}

func (vss *IVssBackupComponents) PrepareForBackup() (HRESULT, *IVSSAsync) {
	var oleIUnknown *ole.IUnknown

	result, _, _ := syscall.Syscall(
		vss.getVTable().prepareForBackup, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&oleIUnknown)),
		0,
	)

	return vss.handleAsyncReturnValue(result, oleIUnknown)
}

func (vss *IVssBackupComponents) handleAsyncReturnValue(result uintptr, oleIUnknown *ole.IUnknown) (HRESULT, *IVSSAsync) {
	if hRes := HRESULT(result); hRes != S_OK {
		return hRes, nil
	} else {

		comInterface, err := queryInterface(oleIUnknown, UIID_IVSS_ASYNC)

		if err != nil {
			// TODO: log here, reconsider return types for better error handling
			return hRes, nil
		}

		iVssAsync := (*IVSSAsync)(unsafe.Pointer(comInterface))

		return hRes, iVssAsync
	}
}

func (vss *IVssBackupComponents) SetBackupState(
	selectComponents bool,
	backupBootableSystemState bool,
	backupType VssBackup,
	partialFileSupport bool,
) HRESULT {

	const TrueValue = 0xffff

	selectComponentsVal := 0
	if selectComponents {
		selectComponentsVal = TrueValue
	}

	backupBootableSystemStateVal := 0
	if backupBootableSystemState {
		backupBootableSystemStateVal = TrueValue
	}

	partialFileSupportVal := 0
	if partialFileSupport {
		partialFileSupportVal = TrueValue
	}

	result, _, _ := syscall.Syscall6(
		vss.getVTable().setBackupState, 5,
		uintptr(unsafe.Pointer(vss)),
		uintptr(selectComponentsVal),
		uintptr(backupBootableSystemStateVal),
		uintptr(backupType),
		uintptr(partialFileSupportVal),
		0,
	)
	return HRESULT(result)
}

func (vss *IVssBackupComponents) DoSnapshotSet() (HRESULT, *IVSSAsync) {
	var oleIUnknown *ole.IUnknown

	result, _, _ := syscall.Syscall(
		vss.getVTable().doSnapshotSet, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&oleIUnknown)),
		0,
	)

	return vss.handleAsyncReturnValue(result, oleIUnknown)
}

func (vss *IVssBackupComponents) DeleteSnapshots(snapshotID ole.GUID) (HRESULT, int32, ole.GUID) {

	var deletedSnapshots int32 = 0
	var nondeletedSnapshotID ole.GUID

	result, _, _ := syscall.Syscall6(
		vss.getVTable().deleteSnapshots, 6,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&snapshotID)),
		uintptr(VSS_OBJECT_SNAPSHOT),
		uintptr(1),
		uintptr(unsafe.Pointer(&deletedSnapshots)),
		uintptr(unsafe.Pointer(&nondeletedSnapshotID)),
	)

	return HRESULT(result), deletedSnapshots, nondeletedSnapshotID
}

func (vss *IVssBackupComponents) GetSnapshotProperties(snapshotID ole.GUID, properties *VssSnapshotProperties) HRESULT {

	result, _, _ := syscall.Syscall(
		vss.getVTable().getSnapshotProperties, 3,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&snapshotID)),
		uintptr(unsafe.Pointer(properties)),
	)

	return HRESULT(result)
}

func (vss *IVssBackupComponents) BackupComplete() (HRESULT, *IVSSAsync) {
	var oleIUnknown *ole.IUnknown

	result, _, _ := syscall.Syscall(
		vss.getVTable().backupComplete, 2,
		uintptr(unsafe.Pointer(vss)),
		uintptr(unsafe.Pointer(&oleIUnknown)),
		0,
	)

	return vss.handleAsyncReturnValue(result, oleIUnknown)
}

func vssFreeSnapshotProperties(properties *VssSnapshotProperties) error {

	if proc, err := findVssProc("VssFreeSnapshotProperties"); err != nil {
		return err
	} else {
		proc.Call(uintptr(unsafe.Pointer(&properties)))
	}

	return nil
}

type VssSnapshotProperties struct {
	snapshotID           ole.GUID
	snapshotSetID        ole.GUID
	snapshotsCount       uint32
	snapshotDeviceObject *uint16
	originalVolumeName   *uint16
	originatingMachine   *uint16
	serviceMachine       *uint16
	exposedName          *uint16
	exposedPath          *uint16
	providerId           ole.GUID
	snapshotAttributes   uint32
	creationTimestamp    uint64
	status               uint
}

func (p *VssSnapshotProperties) GetSnapshotDeviceObject() string {
	return ole.UTF16PtrToString(p.snapshotDeviceObject)
}

var UIID_IVSS_ASYNC = &ole.GUID{
	0x507C37B4, 0xCF5B, 0x4e95,
	[8]byte{0xb0, 0xaf, 0x14, 0xeb, 0x97, 0x67, 0x46, 0x7e},
}

type IVSSAsync struct {
	ole.IUnknown
}

type IVSSAsyncVTable struct {
	ole.IUnknownVtbl
	cancel      uintptr
	wait        uintptr
	queryStatus uintptr
}

const (
	VSS_S_ASYNC_FINISHED = 0x0004230A
)

func (vssAsync *IVSSAsync) getVTable() *IVSSAsyncVTable {
	return (*IVSSAsyncVTable)(unsafe.Pointer(vssAsync.RawVTable))
}

func (vssAsync *IVSSAsync) Wait(millis int64) HRESULT {
	result, _, _ := syscall.Syscall(
		vssAsync.getVTable().wait, 2,
		uintptr(unsafe.Pointer(vssAsync)),
		uintptr(millis),
		0,
	)
	return HRESULT(result)
}

func (vssAsync *IVSSAsync) QueryStatus() (HRESULT, uint32) {
	var state uint32 = 0
	result, _, _ := syscall.Syscall(
		vssAsync.getVTable().queryStatus, 3,
		uintptr(unsafe.Pointer(vssAsync)),
		uintptr(unsafe.Pointer(&state)),
		0,
	)
	return HRESULT(result), state
}

func (vssAsync *IVSSAsync) WaitUntilAsyncFinished(millis int64) error {

	start := time.Now().Unix()

	for {
		hResult := vssAsync.Wait(100)

		if hResult != S_OK {
			// TODO: consider log warning
			continue
		}

		hResult, state := vssAsync.QueryStatus()

		if hResult != S_OK {
			// TODO: consider log warning
			continue
		} else if state == VSS_S_ASYNC_FINISHED {
			return nil
		} else if time.Now().Unix()-start > millis {
			return fmt.Errorf("VSS error: Wait timed out. Waited for more than %d ms.", millis)
		}
	}
}

type VssSnapshot struct {
	iVssBackupComponents *IVssBackupComponents
	snapshotID           ole.GUID
	snapshotProperties   VssSnapshotProperties
	timeoutInMillis      int64
}

func (p *VssSnapshot) Delete() error {
	var err error = nil
	if err = vssFreeSnapshotProperties(&p.snapshotProperties); err != nil {
		return err
	}
	if p.iVssBackupComponents != nil {
		defer p.iVssBackupComponents.Release()

		e := handleAsyncFunctionCall(
			p.iVssBackupComponents.BackupComplete, "BackupComplete", p.timeoutInMillis,
		)
		if e != nil {
			err = e
		}

		if hResult, _, _ := p.iVssBackupComponents.DeleteSnapshots(p.snapshotID); hResult != S_OK {
			err = fmt.Errorf("VSS error: Failed to delete snapshot:  %s (%#x)", hResult.Str(), hResult)
		}
	}

	return err
}

func (p *VssSnapshot) GetSnapshotDeviceObject() string {
	return p.snapshotProperties.GetSnapshotDeviceObject()
}

func NewVssSnapshot(volume string, timeoutInSeconds uint) (VssSnapshot, error) {

	var timeoutInMillis int64 = int64(timeoutInSeconds * 1000)
	vssInstance, err := loadIVssBackupComponentsConstructor()

	if err != nil {
		// TODO: log here
		return VssSnapshot{}, err
	}

	// TODO: consider where to call ole.CoUninitialize()
	err = ole.CoInitialize(0)

	if err != nil {
		// TODO: log here
		return VssSnapshot{}, err
	}

	var oleIUnknown *ole.IUnknown
	result, _, _ := vssInstance.Call(uintptr(unsafe.Pointer(&oleIUnknown)))

	switch HRESULT(result) {
	case S_OK:
	case E_ACCESSDENIED:
		return VssSnapshot{}, fmt.Errorf(
			"VSS error: %s (%#x) The caller does not have sufficient backup privileges or is not an administrator.",
			HRESULT(result).Str(), result,
		)
	default:
		return VssSnapshot{}, fmt.Errorf("VSS error: Failed to create VSS instance:  %s (%#x)", HRESULT(result).Str(), result)
	}

	var comInterface *interface{}
	comInterface, err = queryInterface(oleIUnknown, UUID_IVSS)

	if err != nil {
		// TODO: log here
		return VssSnapshot{}, err
	}

	iVssBackupComponents := (*IVssBackupComponents)(unsafe.Pointer(comInterface))

	if hRes := iVssBackupComponents.InitializeForBackup(); hRes != S_OK {
		return VssSnapshot{}, fmt.Errorf("VSS error: InitializeForBackup() returned %s (%#x)", hRes.Str(), hRes)
	}

	if hRes := iVssBackupComponents.SetContext(VSS_CTX_BACKUP); hRes != S_OK {
		return VssSnapshot{}, fmt.Errorf("VSS error: SetContext() returned %s (%#x)", hRes.Str(), hRes)
	}

	// see https://techcommunity.microsoft.com/t5/Storage-at-Microsoft/What-is-the-difference-between-VSS-Full-Backup-and-VSS-Copy/ba-p/423575
	if hRes := iVssBackupComponents.SetBackupState(false, false, VSS_BT_COPY, false); hRes != S_OK {
		return VssSnapshot{}, fmt.Errorf("VSS error: SetBackupState() returned %s (%#x)", hRes.Str(), hRes)
	}

	err = handleAsyncFunctionCall(iVssBackupComponents.GatherWriterMetadata, "GatherWriterMetadata", timeoutInMillis)
	if err != nil {
		return VssSnapshot{}, err
	}

	// TODO:
	//   What about IVssBackupComponents::AddComponent?
	//   see https://docs.microsoft.com/en-us/windows/desktop/VSS/simple-shadow-copy-creation-for-backup

	hRes, snapshotSetID := iVssBackupComponents.StartSnapshotSet()
	if hRes != S_OK {
		return VssSnapshot{}, fmt.Errorf("VSS error: StartSnapshotSet() returned %s (%#x)", hRes.Str(), hRes)
	}

	if hRes := iVssBackupComponents.AddToSnapshotSet(volume, &snapshotSetID); hRes != S_OK {
		return VssSnapshot{}, fmt.Errorf("VSS error: AddToSnapshotSet() returned %s (%#x)", hRes.Str(), hRes)
	}

	err = handleAsyncFunctionCall(iVssBackupComponents.PrepareForBackup, "PrepareForBackup", timeoutInMillis)
	if err != nil {
		return VssSnapshot{}, err
	}

	err = handleAsyncFunctionCall(iVssBackupComponents.DoSnapshotSet, "DoSnapshotSet", timeoutInMillis)
	if err != nil {
		return VssSnapshot{}, err
	}

	var snapshotProperties VssSnapshotProperties
	if hRes := iVssBackupComponents.GetSnapshotProperties(snapshotSetID, &snapshotProperties); hRes != S_OK {
		return VssSnapshot{}, fmt.Errorf("VSS error: GetSnapshotProperties() returned %s (%#x)", hRes.Str(), hRes)
	}

	/*

		https://microsoft.public.win32.programmer.kernel.narkive.com/aObDj2dD/volume-shadow-copy-backupcomplete-and-vss-e-bad-state

		CreateVSSBackupComponents();
		InitializeForBackup();
		SetBackupState();
		GatherWriterMetadata();
		StartSnapshotSet();
		AddToSnapshotSet();
		PrepareForBackup();
		DoSnapshotSet();
		GetSnapshotProperties();
		<Backup all files>
		VssFreeSnapshotProperties();
		BackupComplete();
	*/

	return VssSnapshot{iVssBackupComponents, snapshotSetID, snapshotProperties, timeoutInMillis}, nil
}

type asyncCallFuncfunc func() (HRESULT, *IVSSAsync)

func handleAsyncFunctionCall(function asyncCallFuncfunc, name string, timeoutInMillis int64) error {

	if hRes, iVssAsync := function(); hRes != S_OK {
		return fmt.Errorf("VSS error: %s() returned %s (%#x)", name, hRes.Str(), hRes)
	} else {

		if iVssAsync == nil {
			return fmt.Errorf("VSS error: %s() returned nil", name)
		}

		err := iVssAsync.WaitUntilAsyncFinished(timeoutInMillis)
		iVssAsync.Release()

		if err != nil {
			return err
		}
	}

	return nil
}

func loadIVssBackupComponentsConstructor() (*windows.LazyProc, error) {

	createInstanceName := "?CreateVssBackupComponents@@YAJPEAPEAVIVssBackupComponents@@@Z"
	if runtime.GOARCH == "386" {
		createInstanceName = "?CreateVssBackupComponents@@YGJPAPAVIVssBackupComponents@@@Z"
	}

	return findVssProc(createInstanceName)
}

func findVssProc(procName string) (*windows.LazyProc, error) {

	vssDll := windows.NewLazySystemDLL("VssApi.dll")
	err := vssDll.Load()

	if err != nil {
		// TODO: log here
		return &windows.LazyProc{}, err
	}

	proc := vssDll.NewProc(procName)
	err = proc.Find()

	if err != nil {
		// TODO: log here
		return &windows.LazyProc{}, err
	}

	return proc, nil
}

func queryInterface(oleIUnknown *ole.IUnknown, guid *ole.GUID) (*interface{}, error) {

	var ivss *interface{}

	result, _, _ := syscall.Syscall(
		oleIUnknown.VTable().QueryInterface,
		3,
		uintptr(unsafe.Pointer(oleIUnknown)),
		uintptr(unsafe.Pointer(guid)),
		uintptr(unsafe.Pointer(&ivss)),
	)

	if result != 0 {
		return nil, fmt.Errorf("QueryInterface failed: %#x", result)
	}

	return ivss, nil
}
