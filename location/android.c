// go:build android

#include <jni.h>
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

static jclass find_class(JNIEnv *env, const char *class_name) {
  jclass clazz = (*env)->FindClass(env, class_name);
  if (clazz == NULL) {
    (*env)->ExceptionClear(env);
    printf("cannot find %s", class_name);
    return NULL;
  }
  return clazz;
}

// Helper function to request permissions from Java
void requestPermissions(JNIEnv *env, jobject activity) {
  // Get the Activity class
  jclass activityClass = (*env)->GetObjectClass(env, activity);

  // Define the permissions you want to request
  const char *permissions[] = {"android.permission.INTERNET",
                               "android.permission.ACCESS_FINE_LOCATION",
                               "android.permission.ACCESS_COARSE_LOCATION"};

  int numPermissions = sizeof(permissions) / sizeof(permissions[0]);

  // Convert to jobjectArray
  jclass stringClass = (*env)->FindClass(env, "java/lang/String");

  jobjectArray permissionsArray =
      (*env)->NewObjectArray(env, numPermissions, stringClass, NULL);

  for (int i = 0; i < numPermissions; i++) {
    jstring permission = (*env)->NewStringUTF(env, permissions[i]);
    (*env)->SetObjectArrayElement(env, permissionsArray, i, permission);
  }

  // Get the requestPermissions method ID
  jmethodID requestPermissionsMethod = (*env)->GetMethodID(
      env, activityClass, "requestPermissions", "([Ljava/lang/String;I)V");

  // Call requestPermissions method
  (*env)->CallVoidMethod(env, activity, requestPermissionsMethod,
                         permissionsArray, 1); // 1 is the request code
}

const char *getCString(uintptr_t jni_env, uintptr_t ctx, jstring str) {
  JNIEnv *env = (JNIEnv *)jni_env;

  const char *chars = (*env)->GetStringUTFChars(env, str, NULL);

  const char *copy = strdup(chars);
  (*env)->ReleaseStringUTFChars(env, str, chars);
  return copy;
}

static jobject getGlobalContext(JNIEnv *env) {
  jclass activityThread = (*env)->FindClass(env, "android/app/ActivityThread");
  jmethodID currentActivityThread =
      (*env)->GetStaticMethodID(env, activityThread, "currentActivityThread",
                                "()Landroid/app/ActivityThread;");
  jobject activityThreadObj = (*env)->CallStaticObjectMethod(
      env, activityThread, currentActivityThread);

  jmethodID getApplication = (*env)->GetMethodID(
      env, activityThread, "getApplication", "()Landroid/app/Application;");
  jobject context =
      (*env)->CallObjectMethod(env, activityThreadObj, getApplication);
  return context;
}

const char *isLocationEnabled(uintptr_t java_vm, uintptr_t jni_env,
                              uintptr_t ctx) {
  JNIEnv *env = (JNIEnv *)jni_env;
  jobject activity = (jobject)ctx;
  requestPermissions(env, activity);
  jobject context = getGlobalContext(env);

  jclass contextClass = find_class(env, "android/app/Application");

  jmethodID getSystemService =
      (*env)->GetMethodID(env, contextClass, "getSystemService",
                          "(Ljava/lang/String;)Ljava/lang/Object;");
  jfieldID locationServiceFieldID = (*env)->GetStaticFieldID(
      env, contextClass, "LOCATION_SERVICE", "Ljava/lang/String;");

  jstring locationServiceString =
      (*env)->GetStaticObjectField(env, contextClass, locationServiceFieldID);

  jobject locationManager = (*env)->CallObjectMethod(
      env, context, getSystemService, locationServiceString);

  jclass locationManagerClass =
      find_class(env, "android/location/LocationManager");

  jmethodID isLocationEnabledID = (*env)->GetMethodID(
      env, locationManagerClass, "isLocationEnabled", "()Z");
  jboolean isLocationEnabledBool =
      (*env)->CallBooleanMethod(env, locationManager, isLocationEnabledID);

  jmethodID getAllProvidersID = (*env)->GetMethodID(
      env, locationManagerClass, "getAllProviders", "()Ljava/util/List;");
  jobject providersList =
      (*env)->CallObjectMethod(env, locationManager, getAllProvidersID);

  jclass listClass = find_class(env, "java/util/List");

  jmethodID listLengthID = (*env)->GetMethodID(env, listClass, "size", "()I");
  jmethodID getFromListID =
      (*env)->GetMethodID(env, listClass, "get", "(I)Ljava/lang/Object;");
  jint providerListLength =
      (*env)->CallIntMethod(env, providersList, listLengthID);

  jstring providerZero =
      (*env)->CallObjectMethod(env, providersList, getFromListID, 0);
  jstring providerOne =
      (*env)->CallObjectMethod(env, providersList, getFromListID, 1);
  jstring providerTwo =
      (*env)->CallObjectMethod(env, providersList, getFromListID, 2);

  jmethodID getLastKnownLocationMethodID =
      (*env)->GetMethodID(env, locationManagerClass, "getLastKnownLocation",
                          "(Ljava/lang/String;)Landroid/location/Location;");
  // fused is generally the best provider
  jobject location = (*env)->CallObjectMethod(
      env, locationManager, getLastKnownLocationMethodID,
      (*env)->NewStringUTF(env, "fused"));

  jclass locationClass = find_class(env, "android/location/Location");

  jmethodID getLatitudeID =
      (*env)->GetMethodID(env, locationClass, "getLatitude", "()D");
  jmethodID getLongitudeID =
      (*env)->GetMethodID(env, locationClass, "getLongitude", "()D");

  jdouble latitude = (*env)->CallDoubleMethod(env, location, getLatitudeID);
  jdouble longitude = (*env)->CallDoubleMethod(env, location, getLongitudeID);

  int providerListLengthInt = (int)providerListLength;
  char *str = "{\"lat\": %f, \"lon\": %f}";
  char *new_str =
      malloc(sizeof(*new_str) * strlen(str) + 1 + 6 + 256 + sizeof(double) * 2);
  sprintf(new_str, str, (double)latitude,
          (double)longitude); // preferrably use fused

  return new_str;
}
