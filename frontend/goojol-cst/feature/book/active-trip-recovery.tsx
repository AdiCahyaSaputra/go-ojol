import { useRouter } from 'expo-router';
import { useEffect, useRef, useState } from 'react';
import { type BookContextValue, useOptionalBook } from '@/feature/book/book-context';
import { fetchActiveTransaction } from '@/feature/book/trip.service';
import { loadStoredActiveTrip, saveStoredActiveTrip } from '@/feature/book/trip-storage';

export function ActiveTripRecovery() {
  const router = useRouter();
  const book = useOptionalBook();
  const checkedRef = useRef(false);

	const [recoverCancelled, setRecoverCancelled] = useState(false);

  const recover = async (book: BookContextValue) => {
    if ((book.transactionId && book.destination) || recoverCancelled) {
      return;
    }

    try {
      const active = await fetchActiveTransaction();

      book.setTransactionId(active.id);
      book.setPickup({
        name: 'Pickup',
        lat: String(active.pickup_lat_long[0]),
        lng: String(active.pickup_lat_long[1]),
      });
      book.setDestination({
        name: 'Destination',
        lat: String(active.destination_lat_long[0]),
        lng: String(active.destination_lat_long[1]),
      });
      if (active.driver) {
        book.setMatchedDriver(active.driver);
      }

      await saveStoredActiveTrip({
        transactionId: active.id,
        pickup: {
          name: 'Pickup',
          lat: String(active.pickup_lat_long[0]),
          lng: String(active.pickup_lat_long[1]),
        },
        destination: {
          name: 'Destination',
          lat: String(active.destination_lat_long[0]),
          lng: String(active.destination_lat_long[1]),
        },
        matchedDriver: active.driver ?? null,
        totalFare: active.total_fare,
      });

      router.push('/book/trip');
    } catch {
      const stored = await loadStoredActiveTrip();
      if (!stored) {
        return;
      }
      try {
        await fetchActiveTransaction();
      } catch {
        await saveStoredActiveTrip(null);
        return;
      }
      book.setTransactionId(stored.transactionId);
      book.setPickup(stored.pickup);
      book.setDestination(stored.destination);
      if (stored.matchedDriver) {
        book.setMatchedDriver(stored.matchedDriver);
      }
      router.push('/book/trip');
    }
  };

  useEffect(() => {
    if (!book || checkedRef.current) {
      return;
    }
    checkedRef.current = true;

		setRecoverCancelled(false);

    void recover(book);

    return () => {
			setRecoverCancelled(true)
    };
  }, [book, router]);

  return null;
}
